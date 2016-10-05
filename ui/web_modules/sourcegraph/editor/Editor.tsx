import * as BlobActions from "sourcegraph/blob/BlobActions";
import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import {URIUtils} from "sourcegraph/core/uri";
import {urlToDefInfo} from "sourcegraph/def/routes";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {EditorService, IEditorOpenedEvent} from "sourcegraph/editor/EditorService";
import {GotoDefinitionWithClickEditorContribution} from "sourcegraph/editor/GotoDefinitionWithClickEditorContribution";
import * as lsp from "sourcegraph/editor/lsp";
import { makeRepoRev } from "sourcegraph/repo";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

const fetch = singleflightFetch(defaultFetch);

function cacheKey(model: monaco.editor.IReadOnlyModel, position: monaco.Position): string | null {
	const word = model.getWordAtPosition(position);
	if (!word) {
		return null;
	}
	return `${model.uri.toString(true)}:${position.lineNumber}:${word.startColumn}:${word.endColumn}`;
}
const hoverCache = new Map<string, any>();
const defCache = new Map<string, any>();

// HACK: don't show "Right-click to view references" on primitive types; if done properly, this
// should be determined by a type property on the hover response.
const refsBlacklist = new Set<string>();
["true", "false"].forEach((bool) => refsBlacklist.add(`const ${bool} untyped bool`));
["bool", "string", "int", "int8", "int16", "int64", "uint", "uint8", "uint16", "uint64", "uintptr", "byte", "rune", "float32", "float64", "complex64", "complex128"].forEach((type) => refsBlacklist.add(`type ${type} ${type}`));
["append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println", "real", "recover"].forEach((builtin) => refsBlacklist.add(`builtin ${builtin}`));
function isPrimitive(contents: any[]): boolean {
	for (let content of contents) {
		if (content instanceof String) {
			if (refsBlacklist.has(content as string)) {
				return true;
			}
		} else if (refsBlacklist.has(content.value)) {
			return true;
		}
	}

	return false;
}

// Editor wraps the Monaco code editor.
export class Editor implements monaco.IDisposable {
	private _editor: monaco.editor.IStandaloneCodeEditor;
	private _editorService: EditorService;
	private _toDispose: monaco.IDisposable[] = [];
	private _initializedModes: Set<string> = new Set();

	constructor(elem: HTMLElement) {
		(global as any).require(["vs/editor/contrib/hover/browser/hoverOperation"], ({HoverOperation}) => {
			HoverOperation.HOVER_TIME = 50;
		});

		// Register services for modes (languages) when new models are added.
		const registerModeProviders = (mode: string) => {
			if (!this._initializedModes.has(mode)) {
				this._toDispose.push(monaco.languages.registerHoverProvider(mode, this));
				this._toDispose.push(monaco.languages.registerDefinitionProvider(mode, this));
				this._toDispose.push(monaco.languages.registerReferenceProvider(mode, this));
				this._initializedModes.add(mode);
			}
		};
		this._toDispose.push(monaco.editor.onDidCreateModel(model => {
			this.disableInterferingModes();
			// HACK: when the editor loads, this will fire twice:
			// - once for the "empty" document (mode = plaintext)
			// - once with the actual mode of the document
			// If we use browser navigation to go from/to this editor, the model
			// will have the correct mode set...but this callback
			// will only fire with the original model (mode = plaintext).
			// (Nor will the onDidChangeModelLanguage callback fire).
			// This hack hardcodes mode providers only for Go, regardless
			// of any state of the editor mode. This way context menu items
			// will *always* appear for Go files, and never for other modes.
			registerModeProviders("go");
		}));

		this._editorService = new EditorService();

		GotoDefinitionWithClickEditorContribution.register(this._editorService);

		this._editor = monaco.editor.create(elem, {
			readOnly: true,
			automaticLayout: true,
			scrollBeyondLastLine: false,
			wrappingColumn: 0,
			fontFamily: code_font_face,
			fontSize: 15,
			lineHeight: 21,
			theme: "vs-dark",
		}, {editorService: this._editorService});

		this._editorService.setEditor(this._editor);

		// Warm up the LSP server immediately when the document loads
		// instead of waiting until the user tries to hover.
		this._editor.onDidChangeModel((e: monaco.editor.IModelChangedEvent) => {
			// HACK: only done for go, since we only registerModeProviders for go.
			if (this._editor.getModel().getModeId() !== "go") {
				return;
			}
			// We modify the name to indicate to our HTTP gateway that this
			// should not be measured as a user triggered action.
			lsp.send(this._editor.getModel(), "textDocument/definition?prepare", {
				textDocument: {uri: e.newModelUrl.toString(true)},
				position: new monaco.Position(0, 0),
			});
		});

		// Remove the "Command Palette" item from the context menu.
		const palette = this._editor.getAction("editor.action.quickCommand");
		if (palette) {
			(palette as any)._shouldShowInContextMenu = false;
			palette.dispose();
		}
		// Don't show context menu for peek view or comments, etc.
		// Also don't show for unsupported languages.
		this._editor.onContextMenu(e => {
			// HACK: This method relies on Monaco private internals.
			const unsupportedLang = this._editor.getModel().getModeId() !== "go";
			// Disable the context menu during chrome onboarding.
			const isOnboarding = location.search.includes("ob=chrome");
			const ident = /.*identifier.*/.exec(e.target.element.className);
			const peekWidget = e.target.detail === "vs.editor.contrib.zoneWidget1";
			if (!ident || peekWidget || unsupportedLang || isOnboarding) {
				(this._editor as any)._contextViewService.hideContextView();
			}
		});

		// Add the "Find External References" item to the context menu.
		this._editor.addAction({
			id: "findExternalReferences",
			label: "Find External References",
			contextMenuGroupId: "1_goto",
			run: (e) => this._findExternalReferences(e.getModel(), e.getPosition()),
		});

		// Rename the "Find All References" action to "Find Local References".
		Object.assign((this._editor.getAction("editor.action.referenceSearch.trigger") || {}) as any, {
			_label: "Find Local References",
		});

		// Monaco overrides the back and forward history commands, so
		// we implement our own here. There currently isn't a way to
		// unbind a default keybinding.
		/* tslint:disable no-bitwise */
		this._editor.addCommand(monaco.KeyCode.LeftArrow | monaco.KeyMod.CtrlCmd, () => {
			global.window.history.back();
		}, "");
		this._editor.addCommand(monaco.KeyCode.RightArrow | monaco.KeyMod.CtrlCmd, () => {
			global.window.history.forward();
		}, "");
		/* tslint:enable no-bitwise */
		this._editor.addCommand(monaco.KeyCode.Home, () => {
			this._editor.revealLine(1);
		}, "");
		this._editor.addCommand(monaco.KeyCode.End, () => {
			this._editor.revealLine(
				this._editor.getModel().getLineCount()
			);
		}, "");
	}

	setInput(uri: monaco.Uri, range?: monaco.IRange): Promise<void> {
		return new Promise((resolve, reject) => {
			this._editorService.openEditor({
				resource: uri,
				options: range ? {selection: range} : undefined,
			}).done(resolve);
		});
	}

	_highlight(startLine: number, startCol?: number, endLine?: number, endCol?: number): void {
		startCol = typeof startCol === "number" ? startCol : this._editor.getModel().getLineMinColumn(startLine);
		endLine = typeof endLine === "number" ? endLine : startLine;
		endCol = typeof endCol === "number" ? endCol : this._editor.getModel().getLineMaxColumn(endLine);
		this._editor.setSelection(new monaco.Range(startLine, startCol, endLine, endCol));
	}

	public trigger(source: string, handlerId: string, payload: any): void {
		this._editor.trigger(source, handlerId, payload);
	}

	// An event emitted when the editor jumps to a new model or position therein.
	public onDidOpenEditor(listener: (e: IEditorOpenedEvent) => void): monaco.IDisposable {
		return this._editorService.onDidOpenEditor(listener);
	}

	provideDefinition(model: monaco.editor.IReadOnlyModel, position: monaco.Position, token: monaco.CancellationToken): monaco.languages.Definition | monaco.Thenable<monaco.languages.Definition | null> {
		const key = cacheKey(model, position);
		if (key) {
			const cacheHit = defCache.get(key);
			if (cacheHit) {
				return Promise.resolve(cacheHit);
			}
		}

		return lsp.send(model, "textDocument/definition", {
			textDocument: {uri: model.uri.toString(true)},
			position: lsp.toPosition(position),
		})
			.then((resp) => resp ? resp.result : null)
			.then((resp: lsp.Location | lsp.Location[] | null) => {
				if (!resp) {
					return null;
				}

				const {repo, rev, path} = URIUtils.repoParams(model.uri);
				EventLogger.logEventForCategory(
					AnalyticsConstants.CATEGORY_DEF,
					AnalyticsConstants.ACTION_CLICK,
					"BlobTokenClicked",
					{ srcRepo: repo, srcRev: rev || "", srcPath: path, language: model.getModeId() }
				);

				const locs: lsp.Location[] = resp instanceof Array ? resp : [resp];
				const translatedLocs: monaco.languages.Location[] = locs
					.filter((loc) => Object.keys(loc).length !== 0)
					.map(lsp.toMonacoLocation);
				if (key) {
					defCache.set(key, translatedLocs);
				}
				return translatedLocs;
			});
	}

	provideHover(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<monaco.languages.Hover> {
		const key = cacheKey(model, position);
		if (key) {
			const cacheHit = hoverCache.get(key);
			if (cacheHit) {
				return Promise.resolve(cacheHit);
			}
		}

		// We have minimal traffic for references. Boost the numbers by sending a small portion of hover traffic to refs.
		// We don't shadow all traffic yet since we are not sure on the resource implications of references yet.
		if (Math.random() < 0.25) {
			lsp.send(model, "textDocument/references", {
				textDocument: {uri: model.uri.toString(true)},
				position: lsp.toPosition(position),
			});
		}

		return lsp.send(model, "textDocument/hover", {
			textDocument: {uri: model.uri.toString(true)},
			position: lsp.toPosition(position),
		})
			.then(resp => {
				if (!resp || !resp.result || !resp.result.contents) {
					return {contents: []}; // if null, strings, whitespace, etc. will show a perpetu-"Loading..." tooltip
				}

				const {repo, rev, path} = URIUtils.repoParams(model.uri);
				EventLogger.logEventForCategory(
					AnalyticsConstants.CATEGORY_DEF,
					AnalyticsConstants.ACTION_HOVER,
					"Hovering",
					{
						repo: repo,
						rev: rev || "",
						path: path,
						language: model.getModeId(),
					}
				);

				let range: monaco.IRange;
				if (resp.result.range) {
					range = lsp.toMonacoRange(resp.result.range);
				} else {
					const word = model.getWordAtPosition(position);
					range = new monaco.Range(position.lineNumber, word ? word.startColumn : position.column, position.lineNumber, word ? word.endColumn : position.column);
				}
				const contents = resp.result.contents instanceof Array ? resp.result.contents : [resp.result.contents];
				if (contents[0].value && contents[0].value.length > 400) {
					contents[0].value = contents[0].value.slice(0, 390) + "...";
				}
				if (!isPrimitive(contents)) {
					contents.push("*Right-click to view references*");
				}
				const hover: monaco.languages.Hover = {
					contents: contents,
					range,
				};
				if (key) {
					hoverCache.set(key, hover);
				}
				return hover;
			});
	}

	provideReferences(model: monaco.editor.IReadOnlyModel, position: monaco.Position, context: monaco.languages.ReferenceContext, token: monaco.CancellationToken): monaco.languages.Location[] | monaco.Thenable<monaco.languages.Location[]> {
		return lsp.send(model, "textDocument/references", {
			textDocument: {uri: model.uri.toString(true)},
			position: lsp.toPosition(position),
			context: {includeDeclaration: false},
		})
			.then((resp) => resp ? resp.result : null)
			.then((resp: lsp.Location | lsp.Location[] | null) => {
				if (!resp) {
					return null;
				}

				const {repo, rev, path} = URIUtils.repoParams(model.uri);
				EventLogger.logEventForCategory(
					AnalyticsConstants.CATEGORY_REFERENCES,
					AnalyticsConstants.ACTION_CLICK,
					"ClickedViewReferences",
					{ repo, rev: rev || "", path }
				);

				const locs: lsp.Location[] = resp instanceof Array ? resp : [resp];
				locs.forEach((l) => {
					l.uri = URIUtils.toRefsDisplayURI(monaco.Uri.parse(l.uri)).toString();
				});
				return locs.map(lsp.toMonacoLocation);
			});
	}

	private	_findExternalReferences(model: monaco.editor.IModel, pos: monaco.IPosition): monaco.Promise<void> {
		const {repo, rev, path} = URIUtils.repoParams(model.uri);
		EventLogger.logEventForCategory(
			AnalyticsConstants.CATEGORY_REFERENCES,
			AnalyticsConstants.ACTION_CLICK,
			"ClickedViewReferences",
			{ repo, rev: rev || "", path }
		);

		const line = pos.lineNumber - 1;
		const col = pos.column - 1;
		return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/hover-info?file=${path}&line=${line}&character=${col}`)
			.then(checkStatus)
			.then(resp => resp.json())
			.catch(err => null)
			.then((resp) => {
				if (resp && (resp as any).def) {
					// TODO(uforic): Remove this when we remove srclib dependency. Fix a special case for golang/go. 
					const def = resp.def;
					if (def.Repo === "github.com/golang/go" && def.Unit && def.Unit.startsWith("github.com/golang/go/src/")) {
						def.Unit = def.Unit.replace("github.com/golang/go/src/", "");
					}
					window.location.href = urlToDefInfo((resp as any).def);
				} else {
					Dispatcher.Stores.dispatch(new BlobActions.Toast("No external references found"));
				}
			});
	}

	// disableInterferingModes disables built-in Monaco features that
	// interfere with Sourcegraph. It retains all modes whose provider
	// is the specified editor, so you must pass it the global editor
	// instance that's currently in use.
	//
	// For example, it disables Monaco's built-in TypeScript language
	// support, so that TypeScript language support comes from
	// Sourcegraph's LSP backend instead.
	//
	// TODO(sqs): If vscode ever becomes more conducive to integrate
	// into our own build system, we can avoid loading these
	// unnecessary things altogether.
	private disableInterferingModes(): void {
		const removeFromLanguageFeatureRegistry = (reg: any) => {
			reg._entries = reg._entries.filter((e) => {
				return e.provider && e.provider === this; // only keep stuff *we* added
			});
		};
		(global as any).require(["vs/editor/common/modes"], (modesModule) => {
			Object.keys(modesModule).forEach((exportedName) => {
				if (exportedName.endsWith("Registry") && modesModule[exportedName]._entries) {
					const reg = modesModule[exportedName];
					removeFromLanguageFeatureRegistry(reg);
					this._toDispose.push(reg.onDidChange(() => removeFromLanguageFeatureRegistry(reg)));
				}
			});
		});
	}

	public layout(): void {
		this._editor.layout();
	}

	public dispose(): void {
		this._editor.dispose();
		this._toDispose.forEach(disposable => {
			disposable.dispose();
		});
	}
}
