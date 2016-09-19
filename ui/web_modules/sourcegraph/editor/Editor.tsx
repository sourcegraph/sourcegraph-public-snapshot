// tslint:disable typedef ordered-imports
import {uriForTreeEntry, treeEntryFromUri} from "sourcegraph/editor/FileModel";
import {EditorService, IEditorOpenedEvent} from "sourcegraph/editor/EditorService";

import { Def } from "sourcegraph/api";
import { makeRepoRev } from "sourcegraph/repo";
import {urlToDefInfo} from "sourcegraph/def/routes";
import {EventLogger} from "sourcegraph/util/EventLogger";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

import {RangeOrPosition} from "sourcegraph/core/rangeOrPosition";

import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import {GotoDefinitionWithClickEditorContribution} from "sourcegraph/editor/GotoDefinitionWithClickEditorContribution";

const fetch = singleflightFetch(defaultFetch);

// Editor wraps the Monaco code editor.
export class Editor implements monaco.IDisposable {
	private _editor: monaco.editor.IStandaloneCodeEditor;
	private _editorService: EditorService;
	private _toDispose: monaco.IDisposable[] = [];
	private _initializedModes: Set<string> = new Set();

	constructor(elem: HTMLElement) {
		// Register services for modes (languages) when new models are added.
		this._toDispose.push(monaco.editor.onDidCreateModel(model => {
			const mode = model.getMode().getId();
			if (!this._initializedModes.has(mode)) {
				this._toDispose.push(monaco.languages.registerHoverProvider(mode, this));
				this._toDispose.push(monaco.languages.registerDefinitionProvider(mode, this));
				if ((window as any).localStorage.monacoReferences) {
					this._toDispose.push(monaco.languages.registerReferenceProvider(mode, this));
				}
				this._initializedModes.add(mode);
			}
		}));

		this._editorService = new EditorService();

		GotoDefinitionWithClickEditorContribution.register(this._editorService);

		this._editor = monaco.editor.create(elem, {
			readOnly: true,
			automaticLayout: true,
			scrollBeyondLastLine: false,
			wrappingColumn: 0,
			fontFamily: code_font_face,
			fontSize: 13,
		}, {editorService: this._editorService});

		this._editorService.setEditor(this._editor);

		// Remove the "Command Palette" item from the context menu.
		const palette = this._editor.getAction("editor.action.quickCommand");
		if (palette) {
			(palette as any)._shouldShowInContextMenu = false;
			palette.dispose();
		}

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
		this._editor.addCommand(monaco.KeyCode.Backspace, () => {
			global.window.history.back();
		}, "");
		this._editor.addCommand(monaco.KeyCode.RightArrow | monaco.KeyMod.CtrlCmd, () => {
			global.window.history.forward();
		}, "");
		this._editor.addCommand(monaco.KeyCode.KEY_S, () => {
			document.body.dispatchEvent(new KeyboardEvent("keydown", { key: "s" }));
		}, "");
		this._editor.addCommand(monaco.KeyCode.KEY_P | monaco.KeyMod.Shift | monaco.KeyMod.CtrlCmd, () => {
			document.body.dispatchEvent(new KeyboardEvent("keydown", { key: "s" }));
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

	provideDefinition(model: monaco.editor.IReadOnlyModel, position: monaco.Position, token: monaco.CancellationToken): monaco.languages.Definition | monaco.Thenable<monaco.languages.Definition> {
		return fetchJumpToDef(model, position).then((resp: JumpToDefResponse) => {
			if (!resp || !resp.Path) {
				return (null as any);
			}

			const uri = monaco.Uri.parse(`sourcegraph://${resp.Path}`);
			if (!uri.fragment) {
				throw new Error(`no uri fragment in uri ${uri.toString()}`);
			}

			const r = RangeOrPosition.parse(uri.fragment.replace(/^L/, ""));
			if (!r) {
				throw new Error(`failed to parse uri fragment ${uri.fragment}`);
			}

			return {
				uri: monaco.Uri.from({scheme: uri.scheme, path: uri.path}),
				range: r.toMonacoRange(),
			};
		});
	}

	provideHover(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<monaco.languages.Hover> {
		return defAtPosition(model, position).then((resp: HoverInfoResponse) => {
			let contents: monaco.MarkedString[] = [];
			if (resp && !resp.Unresolved) {
				if (resp.Title) {
					contents.push(resp.Title);
				}
				contents.push("*Right-click to view all references.*");
			}

			const {repo, rev, path} = treeEntryFromUri(model.uri);
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
	}

	provideReferences(model: monaco.editor.IReadOnlyModel, position: monaco.Position, context: monaco.languages.ReferenceContext, token: monaco.CancellationToken): monaco.languages.Location[] | monaco.Thenable<monaco.languages.Location[]> {
		return refsAtPosition(model, position).then((resp: ReferencesResponse) => {
			const {repo, rev, path} = treeEntryFromUri(model.uri);
			if (!resp) {
				return;
			}

			EventLogger.logEventForCategory(
				AnalyticsConstants.CATEGORY_REFERENCES,
				AnalyticsConstants.ACTION_CLICK,
				"ClickedViewReferences",
				{ repo, rev: rev || "", path }
			);

			const locs: monaco.languages.Location[] = [];
			Object.keys(resp.Locs).forEach((file) => {
				for (let [sl, sc, el, ec] of resp.Locs[file]) {
				locs.push({
					uri: uriForTreeEntry(repo, rev, file),
					range: new monaco.Range(sl + 1, sc + 1, el + 1, ec + 2),
				});
				}
			});
			return locs;
		});
	}

	private	_findExternalReferences(editor: monaco.editor.ICommonCodeEditor): monaco.Promise<void> {
		const model = editor.getModel();
		const pos = editor.getPosition();

		const {repo, rev, path} = treeEntryFromUri(model.uri);
		EventLogger.logEventForCategory(
			AnalyticsConstants.CATEGORY_REFERENCES,
			AnalyticsConstants.ACTION_CLICK,
			"ClickedViewReferences",
			{ repo, rev: rev || "", path }
		);

		return new monaco.Promise<void>(() => {
			defAtPosition(model, pos).then((resp) => {
				if (resp) {
					window.location.href = urlToDefInfo(resp.def);
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
	const {repo, rev, path} = treeEntryFromUri(model.uri);
	return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/hover-info?file=${path}&line=${line}&character=${col}`)
		.then(checkStatus)
		.then(resp => resp.json())
		.catch(err => null);
}

interface JumpToDefResponse {
	Path: string;
}

function fetchJumpToDef(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<JumpToDefResponse> {
	const line = position.lineNumber - 1;
	const col = position.column - 1;
	const {repo, rev, path} = treeEntryFromUri(model.uri);
	return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/jump-def?file=${path}&line=${line}&character=${col}`)
		.then(checkStatus)
		.then(resp => resp.json())
		.catch(err => null);
}

type ReferencesResponse = {
	Locs: {[key: string]: number[][]};
};

function refsAtPosition(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<ReferencesResponse> {
	const line = position.lineNumber - 1;
	const col = position.column - 1;
	const {repo, rev, path} = treeEntryFromUri(model.uri);
	return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/def/dummy/dummy/-/dummy/-/local-refs?file=${path}&line=${line}&character=${col}`)
		.then(checkStatus)
		.then(resp => resp.json())
		.catch(err => null);
}
