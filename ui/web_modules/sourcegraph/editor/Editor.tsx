import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import { URIUtils } from "sourcegraph/core/uri";
import { EditorService, IEditorOpenedEvent } from "sourcegraph/editor/EditorService";
import * as lsp from "sourcegraph/editor/lsp";
import { modes } from "sourcegraph/editor/modes";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { isSupportedMode } from "sourcegraph/util/supportedExtensions";

import "sourcegraph/editor/FindExternalReferencesAction";
import "sourcegraph/editor/GotoDefinitionWithClickEditorContribution";
import "sourcegraph/editor/vscode";

import { CancellationToken } from "vs/base/common/cancellation";
import { KeyCode, KeyMod } from "vs/base/common/keyCodes";
import { IDisposable } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import {IEditorMouseEvent} from "vs/editor/browser/editorBrowser";
import { IEditorConstructionOptions, IStandaloneCodeEditor } from "vs/editor/browser/standalone/standaloneCodeEditor";
import { create as createStandaloneEditor, createModel, onDidCreateModel } from "vs/editor/browser/standalone/standaloneEditor";
import { registerDefinitionProvider, registerHoverProvider, registerReferenceProvider } from "vs/editor/browser/standalone/standaloneLanguages";
import { Position } from "vs/editor/common/core/position";
import { Range } from "vs/editor/common/core/range";
import { IPosition, IRange, IReadOnlyModel, IWordAtPosition } from "vs/editor/common/editorCommon";
import { Definition, Hover, Location, ReferenceContext } from "vs/editor/common/modes";
import { HoverOperation } from "vs/editor/contrib/hover/browser/hoverOperation";

import { MenuId, MenuRegistry } from "vs/platform/actions/common/actions";
import { IEditor } from "vs/platform/editor/common/editor";

function normalisePosition(model: IReadOnlyModel, position: IPosition): IPosition {
	const word = model.getWordAtPosition(position);
	if (!word) {
		return position;
	}
	// We always hover/j2d on the middle of a word. This is so multiple requests for the same word
	// result in a lookup on the same position.
	return {
		lineNumber: position.lineNumber,
		column: Math.floor((word.startColumn + word.endColumn) / 2),
	};
}
function cacheKey(model: IReadOnlyModel, position: IPosition): string {
	return `${model.uri.toString(true)}:${position.lineNumber}:${position.column}`;
}
const hoverCache = new Map<string, Thenable<Hover>>(); // "single-flight" and caching on word boundaries
const defCache = new Map<string, Thenable<Definition | null>>(); // "single-flight" and caching on word boundaries

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
	private _disposed: boolean = false;
	private _initializedModes: Set<string> = new Set();
	private _elementUnderMouse: Element;

	constructor(
		elem: HTMLElement
	) {
		HoverOperation.HOVER_TIME = 200;

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
			modes.forEach(mode => {
				this.registerModeProviders(mode);
			});
		}));

		this._editorService = new EditorService();

		let initialModel = createModel("", "text/plain");
		this._editor = createStandaloneEditor(elem, {
			// If we don't specify an initial model, Monaco will
			// create this one anyway (but it'll try to call
			// window.monaco.editor.createModel, and we don't want to
			// add any implicit dependency on window).
			model: initialModel,

			readOnly: true,
			automaticLayout: true,
			scrollBeyondLastLine: false,
			wrappingColumn: 0,
			fontFamily: code_font_face,
			fontSize: 15,
			lineHeight: 21,
			theme: "vs-dark",
			renderLineHighlight: true,
		}, { editorService: this._editorService });

		// WORKAROUND: Remove the initial model from the configuration to avoid infinite recursion when the config gets updated internally.
		// Reproduce issue by using "Find All References" to open the rift view and then right click again in the code outside of the view.
		delete (this._editor.getRawConfiguration() as IEditorConstructionOptions).model;

		this._editorService.setEditor(this._editor);

		(window as any).ed = this._editor; // for easier debugging via the JS console

		// Don't show context menu for peek view or comments, etc.
		// Also don't show for unsupported languages.
		this._editor.onContextMenu(e => {
			// HACK: This method relies on Monaco private internals.
			const isOnboarding = location.search.includes("ob=chrome");
			const peekWidget = e.target.detail === "vs.editor.contrib.zoneWidget1";
			const c = e.target.element.classList;
			const ignoreToken = c.contains("delimeter") || c.contains("comment") || c.contains("view-line") || (c.length === 1 && c.contains("token"));
			if (ignoreToken || peekWidget || this._editor.getModel() === initialModel || !isSupportedMode(this._editor.getModel().getModeId()) || isOnboarding) {
				(this._editor as any)._contextViewService.hideContextView();
				return;
			}

			const {repo, rev, path} = URIUtils.repoParams(this._editor.getModel().uri);
			AnalyticsConstants.Events.CodeContextMenu_Initiated.logEvent({
					repo: repo,
					rev: rev || "",
					path: path,
					language: this._editor.getModel().getModeId(),
				}
			);

		});

		this._editor.onMouseMove(e => {
			if (e.target.element.classList.contains("token")) {
				this._elementUnderMouse = e.target.element;
			}
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

		// Set the dom readonly property, so keyboard doesn't pop up on mobile.
		const dom = this._editor.getDomNode();
		const input = dom.getElementsByClassName("inputarea");
		if (input.length === 1) {
			input[0].setAttribute("readOnly", "true");
		} else {
			console.error("Didn't set textarea to readOnly");
		}
	}

	// Register services for modes (languages) when new models are added.
	registerModeProviders(mode: string): void {
		if (!this._initializedModes.has(mode)) {
			this._toDispose.push(registerHoverProvider(mode, this));
			this._toDispose.push(registerDefinitionProvider(mode, this));
			this._toDispose.push(registerReferenceProvider(mode, this));
			this._initializedModes.add(mode);
		}
	};

	onLineSelected(listener: (mouseDownEvent: IEditorMouseEvent, mouseUpEvent: IEditorMouseEvent) => void): void {
		let disposeMouseDown = this._editor.onMouseDown(mouseDownEvent => {
			let disposeMouseUp = this._editor.onMouseUp(function(mouseUpEvent: IEditorMouseEvent): void {
				listener(mouseDownEvent, mouseUpEvent);
				disposeMouseUp.dispose();
			});
		});

		this._toDispose.push(disposeMouseDown);
	}

	setInput(uri: URI, range?: IRange): Promise<IEditor> {
		return new Promise<IEditor>((resolve, reject) => {
			this._editorService.openEditor({
				resource: uri,
				options: range ? { selection: range } : undefined,
			})
			.done(resolve, reject);
		});
	}

	public setSelection(range: IRange): void {
		this._editor.setSelection(range);
	}

	public getSelection(): any {
		return this._editor.getSelection();
	}

	public trigger(source: string, handlerId: string, payload: any): void {
		this._editor.trigger(source, handlerId, payload);
	}

	// An event emitted when the editor jumps to a new model or position therein.
	public onDidOpenEditor(listener: (e: IEditorOpenedEvent) => void): IDisposable {
		return this._editorService.onDidOpenEditor(listener);
	}

	provideDefinition(model: IReadOnlyModel, origPosition: Position, token: CancellationToken): Thenable<Definition | null> {
		const position = normalisePosition(model, origPosition);
		const key = cacheKey(model, position);
		const cached = defCache.get(key);
		if (cached) {
			return cached;
		}

		const flight = lsp.send(model, "textDocument/definition", {
			textDocument: { uri: URIUtils.fromRefsDisplayURIMaybe(model.uri).toString(true) },
			position: lsp.toPosition(position),
		})
			.then((resp) => resp ? resp.result : null)
			.then((resp: lsp.Location | lsp.Location[] | null) => {
				if (!resp) {
					return null;
				}

				const locs: lsp.Location[] = resp instanceof Array ? resp : [resp];
				const translatedLocs: Location[] = locs
					.filter((loc) => Object.keys(loc).length !== 0)
					.map(lsp.toMonacoLocation);

				if (this._disposed) {
					defCache.delete(key);
					return null; // need to return null, otherwise vscode errors internally
				}
				return translatedLocs;
			});

		defCache.set(key, flight);

		return flight;
	}

	provideHover(model: IReadOnlyModel, origPosition: Position): Thenable<Hover> {
		const position = normalisePosition(model, origPosition);
		const word = model.getWordAtPosition(position);
		const key = cacheKey(model, position);
		const cached = hoverCache.get(key);
		if (cached) {
			return cached.then(hover => {
				if (hover.contents && hover.contents.length > 0) {
					this.setTokenCursor(word);
				}
				return hover;
			});
		}

		const flight = lsp.send(model, "textDocument/hover", {
			textDocument: { uri: URIUtils.fromRefsDisplayURIMaybe(model.uri).toString(true) },
			position: lsp.toPosition(position),
		})
			.then(resp => {
				if (!resp || !resp.result || !resp.result.contents || resp.result.contents.length === 0) {
					return { contents: [] }; // if null, strings, whitespace, etc. will show a perpetu-"Loading..." tooltip
				}

				const {repo, rev, path} = URIUtils.repoParams(model.uri);
				AnalyticsConstants.Events.CodeToken_Hovered.logEvent({
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
					range = new Range(position.lineNumber, word ? word.startColumn : position.column, position.lineNumber, word ? word.endColumn : position.column);
				}
				const contents = resp.result.contents instanceof Array ? resp.result.contents : [resp.result.contents];
				for (const c of contents) {
					if (c.value && c.value.length > 300) {
						c.value = c.value.slice(0, 300) + "...";
					}
				}

				// HACK: temporarily render Markdown hover strings as
				// sans-serif plain text to work around the Markdown
				// rendering issue in
				// https://github.com/sourcegraph/sourcegraph/issues/1947
				// where prose is rendered as monospace. For some
				// reason, this actually renders Markdown correctly
				// (code is monospace, prose is sans-serif), whereas
				// without this, those are rendered in the opposite
				// ways (code is sans-serif, prose is monospace).
				for (let i = 0; i < contents.length; i++) {
					if (contents[i].language === "markdown") {
						contents[i] = contents[i].value;
					}
				}
				// END HACK

				if (!isPrimitive(contents)) {
					contents.push("*Right-click to view references*");
				}
				const hover: Hover = {
					contents: contents,
					range,
				};
				this.setTokenCursor(word);
				return hover;
			});

		hoverCache.set(key, flight);

		return flight;
	}

	provideReferences(model: IReadOnlyModel, position: Position, context: ReferenceContext, token: CancellationToken): Location[] | Thenable<Location[]> {
		return lsp.send(model, "textDocument/references", {
			textDocument: { uri: model.uri.toString(true) },
			position: lsp.toPosition(position),
			context: { includeDeclaration: false },
		})
			.then((resp) => resp ? resp.result : null)
			.then((resp: lsp.Location | lsp.Location[] | null) => {
				if (!resp || Object.keys(resp).length === 0) {
					return null;
				}

				const {repo, rev, path} = URIUtils.repoParams(model.uri);
				AnalyticsConstants.Events.CodeReferences_Viewed.logEvent({ repo, rev: rev || "", path });

				const locs: lsp.Location[] = resp instanceof Array ? resp : [resp];
				locs.forEach((l) => {
					l.uri = URIUtils.toRefsDisplayURI(URI.parse(l.uri)).toString();
				});
				return locs.map(lsp.toMonacoLocation);
			});
	}

	private setTokenCursor(word: IWordAtPosition): void {
		// model.getWordAtPosition can return null.
		if (!word) {
			return;
		}
		const el = (this._elementUnderMouse as any);
		// Make sure the mouse is still under the target word.
		if (el && el.textContent === word.word) {
			// Ensure tokens that don't identifier class but do have hover info get a pointer cursor.
			el.style.cursor = "pointer";
		}
	}

	public layout(): void {
		this._editor.layout();
	}

	public dispose(): void {
		this._disposed = true;
		this._editor.dispose();
		this._toDispose.forEach(disposable => {
			disposable.dispose();
		});
	}
}
