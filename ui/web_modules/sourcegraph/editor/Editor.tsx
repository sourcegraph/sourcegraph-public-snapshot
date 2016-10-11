import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import {URIUtils} from "sourcegraph/core/uri";
import {EditorService, IEditorOpenedEvent} from "sourcegraph/editor/EditorService";
import * as lsp from "sourcegraph/editor/lsp";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import {isSupportedMode, typescriptSupported} from "sourcegraph/util/supportedExtensions";

import "sourcegraph/editor/FindExternalReferencesAction";
import "sourcegraph/editor/GotoDefinitionWithClickEditorContribution";
import "sourcegraph/editor/vscode";

import {CancellationToken} from "vs/base/common/cancellation";
import {KeyCode, KeyMod} from "vs/base/common/keyCodes";
import {IDisposable} from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import {IStandaloneCodeEditor} from "vs/editor/browser/standalone/standaloneCodeEditor";
import {create as createStandaloneEditor, createModel, onDidCreateModel} from "vs/editor/browser/standalone/standaloneEditor";
import {registerDefinitionProvider, registerHoverProvider, registerReferenceProvider} from "vs/editor/browser/standalone/standaloneLanguages";
import {Position} from "vs/editor/common/core/position";
import {Range} from "vs/editor/common/core/range";
import {IPosition, IRange, IReadOnlyModel} from "vs/editor/common/editorCommon";
import {Definition, Hover, Location, ReferenceContext} from "vs/editor/common/modes";
import {HoverOperation} from "vs/editor/contrib/hover/browser/hoverOperation";

import {MenuId, MenuRegistry} from "vs/platform/actions/common/actions";

function cacheKey(model: IReadOnlyModel, position: IPosition): string | null {
	const word = model.getWordAtPosition(position);
	if (!word) {
		return null;
	}
	return `${model.uri.toString(true)}:${position.lineNumber}:${word.startColumn}:${word.endColumn}`;
}
const hoverCache = new Map<string, any>();
const hoverFlights = new Map<string, Thenable<Hover>>(); // "single-flight" on word boundaries
const defCache = new Map<string, any>();
const defFlights = new Map<string, Thenable<Definition | null>>(); // "single-flight" on word boundaries

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
export class Editor implements IDisposable {
	private _editor: IStandaloneCodeEditor;
	private _editorService: EditorService;
	private _toDispose: IDisposable[] = [];
	private _initializedModes: Set<string> = new Set();
	private _mouseIsOverIdentifier: boolean = false;

	constructor(
		elem: HTMLElement
	) {
		HoverOperation.HOVER_TIME = 50;

		// Register services for modes (languages) when new models are added.
		const registerModeProviders = (mode: string) => {
			if (!this._initializedModes.has(mode)) {
				this._toDispose.push(registerHoverProvider(mode, this));
				this._toDispose.push(registerDefinitionProvider(mode, this));
				this._toDispose.push(registerReferenceProvider(mode, this));
				this._initializedModes.add(mode);
			}
		};
		this._toDispose.push(onDidCreateModel(model => {
			// HACK: when the editor loads, this will fire twice:
			// - once for the "empty" document (mode = plaintext)
			// - once with the actual mode of the document
			// If we use browser navigation to go from/to this editor, the model
			// will have the correct mode set...but this callback
			// will only fire with the original model (mode = plaintext).
			// (Nor will the onDidChangeModelLanguage callback fire).
			// This hack hardcodes mode providers only for Go, regardless
			// of any state of the editor mode. This way context menu items
			// will *always* appear for files with the extensions below.
			registerModeProviders("go");
			if (typescriptSupported()) {
				registerModeProviders("typescript");
			}
		}));

		this._editorService = new EditorService();

		this._editor = createStandaloneEditor(elem, {
			// If we don't specify an initial model, Monaco will
			// create this one anyway (but it'll try to call
			// window.monaco.editor.createModel, and we don't want to
			// add any implicit dependency on window).
			model: createModel("", "text/plain"),

			readOnly: true,
			automaticLayout: true,
			scrollBeyondLastLine: false,
			wrappingColumn: 0,
			fontFamily: code_font_face,
			fontSize: 15,
			lineHeight: 21,
			theme: "vs-dark",
			renderLineHighlight: true,
		}, {editorService: this._editorService});

		this._editorService.setEditor(this._editor);

		(window as any).ed = this._editor; // for easier debugging via the JS console

		// Don't show context menu for peek view or comments, etc.
		// Also don't show for unsupported languages.
		this._editor.onContextMenu(e => {
			// HACK: This method relies on Monaco private internals.
			const isOnboarding = location.search.includes("ob=chrome");
			const ident = /.*identifier.*/.exec(e.target.element.className);
			const peekWidget = e.target.detail === "vs.editor.contrib.zoneWidget1";
			if (!ident || peekWidget || !isSupportedMode(this._editor.getModel().getModeId()) || isOnboarding) {
				(this._editor as any)._contextViewService.hideContextView();
			}
		});

		this._editor.onMouseMove(e => {
			this._mouseIsOverIdentifier = Boolean(/.*identifier.*/.exec(e.target.element.className));
		});

		// Rename the "Find All References" action to "Find Local References".
		Object.assign((this._editor.getAction("editor.action.referenceSearch.trigger") || {}) as any, {
			_label: "Find Local References",
		});

		// Monaco overrides the back and forward history commands, so
		// we implement our own here. There currently isn't a way to
		// unbind a default keybinding.
		/* tslint:disable no-bitwise */
		this._editor.addCommand(KeyCode.LeftArrow | KeyMod.CtrlCmd, () => {
			global.window.history.back();
		}, "");
		this._editor.addCommand(KeyCode.RightArrow | KeyMod.CtrlCmd, () => {
			global.window.history.forward();
		}, "");
		/* tslint:enable no-bitwise */
		this._editor.addCommand(KeyCode.Home, () => {
			this._editor.revealLine(1);
		}, "");
		this._editor.addCommand(KeyCode.End, () => {
			this._editor.revealLine(
				this._editor.getModel().getLineCount()
			);
		}, "");

		let editorMenuItems = MenuRegistry.getMenuItems(MenuId.EditorContext);
		let commandOrder = {
			"editor.action.referenceSearch.trigger": 1.1,
			"editor.action.previewDeclaration": 1.2,
			"editor.action.goToDeclaration": 1.3,
		};
		for (let item of editorMenuItems) {
			item.order = commandOrder[item.command.id] || item.order;
			// HACK: VSCode doesn't have a clean API for removing context menu items
			// we don't want. The Copy action shows up always so remove it manually.
			if (item.command.id === "editor.action.clipboardCopyAction") {
				const idx = editorMenuItems.indexOf(item);
				if (idx >= 0) {
					editorMenuItems.splice(idx, 1);
				}
			}
		}
	}

	setInput(uri: URI, range?: IRange): Promise<void> {
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
		this._editor.setSelection(new Range(startLine, startCol, endLine, endCol));
	}

	public trigger(source: string, handlerId: string, payload: any): void {
		this._editor.trigger(source, handlerId, payload);
	}

	// An event emitted when the editor jumps to a new model or position therein.
	public onDidOpenEditor(listener: (e: IEditorOpenedEvent) => void): IDisposable {
		return this._editorService.onDidOpenEditor(listener);
	}

	provideDefinition(model: IReadOnlyModel, position: Position, token: CancellationToken): Thenable<Definition | null> {
		const key = cacheKey(model, position);
		if (key) {
			const cacheHit = defCache.get(key);
			if (cacheHit) {
				return Promise.resolve(cacheHit);
			}
			const inFlight = defFlights.get(key);
			if (inFlight) {
				return inFlight;
			}
		}

		const flight = lsp.send(model, "textDocument/definition", {
			textDocument: {uri: URIUtils.fromRefsDisplayURIMaybe(model.uri).toString(true)},
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
				const translatedLocs: Location[] = locs
					.filter((loc) => Object.keys(loc).length !== 0)
					.map(lsp.toMonacoLocation);
				if (key) {
					defCache.set(key, translatedLocs);
					defFlights.delete(key);
				}
				return translatedLocs;
			});

		if (key) {
			defFlights.set(key, flight);
		}

		return flight;
	}

	provideHover(model: IReadOnlyModel, position: Position): Thenable<Hover> {
		const key = cacheKey(model, position);
		if (key) {
			const cacheHit = hoverCache.get(key);
			if (cacheHit) {
				return Promise.resolve(cacheHit);
			}
			const inFlight = hoverFlights.get(key);
			if (inFlight) {
				return inFlight;
			}
		}

		// HACK(john): VSCode sends a hover request whenever your cursor moves over any token of
		// buffer content. We should remove this and submit a patch to VSCode to short-circuit
		// hover requests for certain tokens.
		if (!this._mouseIsOverIdentifier) {
			return Promise.resolve({contents: []});
		}

		const flight = lsp.send(model, "textDocument/hover", {
			textDocument: {uri: URIUtils.fromRefsDisplayURIMaybe(model.uri).toString(true)},
			position: lsp.toPosition(position),
		})
			.then(resp => {
				if (!resp || !resp.result || !resp.result.contents || resp.result.contents.length === 0) {
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

				let range: IRange;
				if (resp.result.range) {
					range = lsp.toMonacoRange(resp.result.range);
				} else {
					const word = model.getWordAtPosition(position);
					range = new Range(position.lineNumber, word ? word.startColumn : position.column, position.lineNumber, word ? word.endColumn : position.column);
				}
				const contents = resp.result.contents instanceof Array ? resp.result.contents : [resp.result.contents];
				if (contents[0].value && contents[0].value.length > 400) {
					contents[0].value = contents[0].value.slice(0, 390) + "...";
				}
				if (!isPrimitive(contents)) {
					contents.push("*Right-click to view references*");
				}
				const hover: Hover = {
					contents: contents,
					range,
				};
				if (key) {
					hoverCache.set(key, hover);
					hoverFlights.delete(key);
				}
				return hover;
			});

		if (key) {
			hoverFlights.set(key, flight);
		}

		return flight;
	}

	provideReferences(model: IReadOnlyModel, position: Position, context: ReferenceContext, token: CancellationToken): Location[] | Thenable<Location[]> {
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
					l.uri = URIUtils.toRefsDisplayURI(URI.parse(l.uri)).toString();
				});
				return locs.map(lsp.toMonacoLocation);
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
