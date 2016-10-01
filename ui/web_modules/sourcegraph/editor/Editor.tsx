// tslint:disable typedef ordered-imports
import {URI} from "sourcegraph/core/uri";
import {EditorService, IEditorOpenedEvent} from "sourcegraph/editor/EditorService";

import * as lsp from "sourcegraph/editor/lsp";
import { Def } from "sourcegraph/api";
import { makeRepoRev } from "sourcegraph/repo";
import {urlToDefInfo} from "sourcegraph/def/routes";
import {EventLogger} from "sourcegraph/util/EventLogger";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import {GotoDefinitionWithClickEditorContribution} from "sourcegraph/editor/GotoDefinitionWithClickEditorContribution";

const fetch = singleflightFetch(defaultFetch);

const neverPromise = new Promise(() => null); // never resolved

const promiseRaceResolvedOnly = (a, b: Promise<any>): Promise<any> => {
	// Promise.race returns the first promise that is either resolved
	// or rejected. We want a funciton that returns the first promise
	// that is resolved. If both are rejected, then return one of the
	// rejections so that the promise eventually terminates.
	return new Promise((resolve, reject) => {
		let waitingForOther = false;
		let done = false;
		const resolve2 = (v) => {
			if (!v) {
				waitingForOther = true;
			} else if (!done) {
				resolve(v);
				done = true;
			}
		};
		const reject2 = (err) => {
			if (!waitingForOther) {
				waitingForOther = true;
			} else if (!done) {
				reject(err);
				done = true;
			}
		};
		a.then(resolve2, reject2);
		b.then(resolve2, reject2);
	});
};

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
			const mode = model.getMode().getId();
			registerModeProviders(mode);
		}));
		this._toDispose.push(monaco.editor.onDidChangeModelLanguage(e => {
			// In case the mode doesn't get loaded async until after
			// onDidCreateModel, try to register again.
			registerModeProviders(e.model.getModeId());
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

		// Remove the "Command Palette" item from the context menu.
		const palette = this._editor.getAction("editor.action.quickCommand");
		if (palette) {
			(palette as any)._shouldShowInContextMenu = false;
			palette.dispose();
		}
		// Don't show context menu for peek view or comments, etc.
		this._editor.onContextMenu(e => {
			// HACK: This method relies on Monaco private internals.
			const ident = /.*identifier.*/.exec(e.target.element.className);
			const peekWidget = e.target.detail === "vs.editor.contrib.zoneWidget1";
			if (!ident || peekWidget) {
				(this._editor as any)._contextViewService.hideContextView();
			}
		});

		// Add the "Find External References" item to the context menu.
		this._editor.addAction({
			id: "findExternalReferences",
			label: "Find External References",
			contextMenuGroupId: "1_goto",
			run: (e) => this._findExternalReferences(e),
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
		const xlangRes = (window.localStorage as any).langpOnly ? neverPromise : lsp.send(model, "textDocument/definition", {
			textDocument: {uri: model.uri.toString(true)},
			position: lsp.toPosition(position),
		})
			.then((resp) => resp ? resp.result : null)
			.then((resp: lsp.Location | lsp.Location[] | null) => {
				if (!resp) {
					return null;
				}
				const locs: lsp.Location[] = resp instanceof Array ? resp : [resp];
				return locs.map(lsp.toMonacoLocation);
			});

		const langpRes = (window.localStorage as any).xlangOnly ? neverPromise : fetchJumpToDef(model, position).then((loc: lsp.Location) => loc && loc.uri ? lsp.toMonacoLocation(loc) : null);

		// Return whichever result (langp or xlang) arrives first. The
		// feature flags localStorage.{langp,xlang}Only cause only
		// results from that source to be used (but requests are still
		// sent off to both).
		return promiseRaceResolvedOnly(xlangRes, langpRes as Promise<any>);
	}

	provideHover(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<monaco.languages.Hover> {
		const xlangRes = (window.localStorage as any).langpOnly ? neverPromise : lsp.send(model, "textDocument/hover", {
			textDocument: {uri: model.uri.toString(true)},
			position: lsp.toPosition(position),
		})
			.then(resp => {
				if (!resp || !resp.result || !resp.result.contents) {
					return {contents: []}; // if null, strings, whitespace, etc. will show a perpetu-"Loading..." tooltip
				}

				let range: monaco.IRange;
				if (resp.result.range) {
					range = lsp.toMonacoRange(resp.result.range);
				} else {
					const word = model.getWordAtPosition(position);
					range = new monaco.Range(position.lineNumber, word ? word.startColumn : position.column, position.lineNumber, word ? word.endColumn : position.column);
				}
				return {
					contents: resp.result.contents,
					range,
				};
			});

		const langpRes = (window.localStorage as any).xlangOnly ? neverPromise : defAtPosition(model, position).then((resp: HoverInfoResponse) => {
			let contents: monaco.MarkedString[] = [];
			if (!resp) {
				// No-op.
			} else if (resp && !(resp as any).Unresolved) {
				if ((resp as any).Title) {
					contents.push((resp as any).Title);
				}
				contents.push("*Right-click to view all references.*");
			}

			const {repo, rev, path} = URI.repoParams(model.uri);
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

			const word = model.getWordAtPosition(position) || new monaco.Range(position.lineNumber, position.column, position.lineNumber, position.column);
			return {
				range: new monaco.Range(position.lineNumber, word.startColumn, position.lineNumber, word.endColumn),
				contents: contents,
			};
		});

		return promiseRaceResolvedOnly(xlangRes, langpRes as Promise<any>);
	}

	provideReferences(model: monaco.editor.IReadOnlyModel, position: monaco.Position, context: monaco.languages.ReferenceContext, token: monaco.CancellationToken): monaco.languages.Location[] | monaco.Thenable<monaco.languages.Location[]> {
		const xlangRes = (window.localStorage as any).langpOnly ? neverPromise : lsp.send(model, "textDocument/references", {
			textDocument: {uri: model.uri.toString(true)},
			position: lsp.toPosition(position),
			context: {includeDeclaration: false},
		})
			.then((resp) => resp ? resp.result : null)
			.then((resp: lsp.Location | lsp.Location[] | null) => {
				if (!resp) {
					return null;
				}
				const locs: lsp.Location[] = resp instanceof Array ? resp : [resp];
				locs.forEach((l) => {
					l.uri = URI.toRefsDisplayURI(monaco.Uri.parse(l.uri)).toString();
				});
				return locs.map(lsp.toMonacoLocation);
			});

		const langpRes = (window.localStorage as any).xlangOnly ? neverPromise : refsAtPosition(model, position).then((resp: ReferencesResponse) => {
			const {repo, rev, path} = URI.repoParams(model.uri);
			if (!resp) {
				return;
			}

			EventLogger.logEventForCategory(
				AnalyticsConstants.CATEGORY_REFERENCES,
				AnalyticsConstants.ACTION_CLICK,
				"ClickedViewReferences",
				{ repo, rev: rev || "", path }
			);

			return resp.Locs.map(lsp.toMonacoLocation);
		});

		return promiseRaceResolvedOnly(xlangRes, langpRes as Promise<any>);
	}

	private	_findExternalReferences(editor: monaco.editor.ICommonCodeEditor): monaco.Promise<void> {
		const model = editor.getModel();
		const pos = editor.getPosition();

		const {repo, rev, path} = URI.repoParams(model.uri);
		EventLogger.logEventForCategory(
			AnalyticsConstants.CATEGORY_REFERENCES,
			AnalyticsConstants.ACTION_CLICK,
			"ClickedViewReferences",
			{ repo, rev: rev || "", path }
		);

		return new monaco.Promise<void>(() => {
			defAtPosition(model, pos).then((resp) => {
				if (resp && (resp as any).def) {
					window.location.href = urlToDefInfo((resp as any).def);
				} else {
					console.error("No def landing URL found.", resp);
				}
			});
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

interface HoverInfoResponse {
	def: Def;
	Title?: string;
	Unresolved?: boolean;
}

function defAtPosition(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<HoverInfoResponse> {
	const line = position.lineNumber - 1;
	const col = position.column - 1;
	const {repo, rev, path} = URI.repoParams(model.uri);
	return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/hover-info?file=${path}&line=${line}&character=${col}`)
		.then(checkStatus)
		.then(resp => resp.json())
		.catch(err => null);
}

interface JumpToDefResponse {
	Path: string;
}

function fetchJumpToDef(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<lsp.Location> {
	const line = position.lineNumber - 1;
	const col = position.column - 1;
	const {repo, rev, path} = URI.repoParams(model.uri);
	return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/jump-def?file=${path}&line=${line}&character=${col}`)
		.then(checkStatus)
		.then(resp => resp.json())
		.catch(err => null);
}

type ReferencesResponse = {
	Locs: lsp.Location[];
};

function refsAtPosition(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<ReferencesResponse> {
	const line = position.lineNumber - 1;
	const col = position.column - 1;
	const {repo, rev, path} = URI.repoParams(model.uri);
	return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/def/dummy/dummy/-/dummy/-/local-refs?file=${path}&line=${line}&character=${col}`)
		.then(checkStatus)
		.then(resp => resp.json())
		.catch(err => null);
}
