// tslint:disable typedef ordered-imports
import {EventLogger} from "sourcegraph/util/EventLogger";
import {URI} from "sourcegraph/core/uri";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import * as debounce from "lodash/debounce";

// tslint:disable typedef ordered-imports member-ordering
import {IEditorService} from "sourcegraph/editor/EditorService";

// IWordAtPositionWithLine lets us distinguish between two of the same
// words at the same start column on separate lines.
interface IWordAtPositionWithLine extends monaco.editor.IWordAtPosition {
	lineNumber: number; // The line number where the word starts.
}

export class GotoDefinitionWithClickEditorContribution implements monaco.editor.IEditorContribution {
	// register registers this contribution and sets the EditorService.
	static register(editorService: IEditorService): Promise<void> {
		GotoDefinitionWithClickEditorContribution.editorService = editorService;
		return new Promise((resolve, reject) => {
			(global as any).require(["vs/editor/browser/editorBrowserExtensions"], (m) => {
				let f = m.editorContribution /* monaco-editor 0.6.1 */ || m.EditorBrowserRegistry.registerEditorContribution /* =~ 0.5 */;
				f(GotoDefinitionWithClickEditorContribution);
				resolve();
			});
		});
	}

	private static ID = "editor.contrib.gotodefinitionwithclick";

	private static editorService: IEditorService;

	private editor: monaco.editor.ICodeEditor;
	private toUnhook: monaco.IDisposable[] = [];
	private decorations: string[] = [];
	private selectedDefDecoration: string[] = [];
	private currentWordUnderMouse: IWordAtPositionWithLine | null;
	private lastMouseMoveEvent: monaco.editor.IEditorMouseEvent | null;
	private findDefinitionDebounced: (target: monaco.editor.IMouseTarget, word: monaco.editor.IWordAtPosition) => void;
	private mouseLine: number;
	private mouseColumn: number;

	constructor(
		editor: monaco.editor.ICodeEditor,
	) {
		this.editor = editor;

		this.toUnhook.push(this.editor.onMouseUp((e: monaco.editor.IEditorMouseEvent) => this.onEditorMouseUp(e)));
		this.toUnhook.push(this.editor.onMouseMove((e: monaco.editor.IEditorMouseEvent) => this.onEditorMouseMove(e)));

		this.toUnhook.push(this.editor.onDidChangeCursorSelection((e) => this.onDidChangeCursorSelection(e)));
		this.toUnhook.push(this.editor.onDidChangeModel((e) => this.resetHandler()));
		this.toUnhook.push(this.editor.onDidChangeModelContent(() => this.resetHandler()));
		this.toUnhook.push(this.editor.onMouseMove((e) => {
			if (!e.target.position) {
				return;
			}
			this.mouseLine = e.target.position.lineNumber;
			this.mouseColumn = e.target.position.column;
		}));
		this.toUnhook.push(this.editor.onDidScrollChange((e) => {
			if (e.scrollTopChanged || e.scrollLeftChanged) {
				this.resetHandler();
			}
		}));

		this.findDefinitionDebounced = debounce(this.findDefinitionDebounced_, 150, { leading: true, trailing: true });
	}

	private onDidChangeCursorSelection(e: monaco.editor.ICursorSelectionChangedEvent): void {
		if (e.selection && e.selection.startColumn !== e.selection.endColumn) {
			this.resetHandler(); // immediately stop this feature if the user starts to select (https://github.com/Microsoft/vscode/issues/7827)
		}

		// After the selection is changed check to see if the new current selection
		// has landed on a definition. If so, highlight it.
		this.highlightDefinitionAtSelection(e.selection);
	}

	private onEditorMouseMove(mouseEvent: monaco.editor.IEditorMouseEvent): void {
		if (mouseEvent.target.type === monaco.editor.MouseTargetType.UNKNOWN) {
			// Occurs occasionally when mousing over syntax-highlighted tokens. Must ignore or
			// else the decorations will erroneously be removed.
			return;
		}

		this.startFindDefinition(mouseEvent);
		this.lastMouseMoveEvent = mouseEvent;
	}

	private startFindDefinition(mouseEvent: monaco.editor.IEditorMouseEvent, withKey?: monaco.IKeyboardEvent): void {
		if (!this.isEnabled(mouseEvent)) {
			this.currentWordUnderMouse = null;
			this.removeDecorations();
			return;
		}

		// Find word at mouse position
		let position = mouseEvent.target.position;
		const wordAtPos = position ? this.editor.getModel().getWordAtPosition(position) : null;
		if (!wordAtPos) {
			this.currentWordUnderMouse = null;
			this.removeDecorations();
			return;
		}
		const word = Object.assign(wordAtPos, {lineNumber: position.lineNumber});
		if (word.endColumn === position.column) {
			// The end column of a word is the position AFTER the last character in
			// the word. Prevent this.currentWordUserMouse from being set while
			// hovering over a character outside of the word.
			return;
		}

		// Return early if word at position is still the same
		if (this.currentWordUnderMouse && this.currentWordUnderMouse.lineNumber === word.lineNumber && this.currentWordUnderMouse.startColumn === word.startColumn && this.currentWordUnderMouse.endColumn === word.endColumn && this.currentWordUnderMouse.word === word.word) {
			return;
		}

		this.currentWordUnderMouse = word;
		this.findDefinitionDebounced(mouseEvent.target, word);
	}

	private findDefinitionDebounced_(target: monaco.editor.IMouseTarget, word: monaco.editor.IWordAtPosition): void {
		this.findDefinition(target.position).then(results => {
			if (!results || !results.length) {
				this.removeDecorations();
				return;
			}

			// If the mouse isn't currently over the word we just fetched, don't highlight it.
			if (this.mouseLine !== target.position.lineNumber || this.mouseColumn < word.startColumn || word.endColumn < this.mouseColumn) {
				return;
			}
			this.addDecoration(
				{
					startLineNumber: target.position.lineNumber,
					startColumn: word.startColumn,
					endLineNumber: target.position.lineNumber,
					endColumn: word.endColumn,
				},
				results.length > 1 ? `Click to show the ${results.length} definitions found.` : undefined
			);
		});
	}

	private highlightDefinitionAtSelection(selection: monaco.Selection) {
		let position = ({
			lineNumber: selection.startLineNumber,
			column: selection.startColumn,
		});
		this.findDefinition(position).then(results => {
			if (!results || !results.length) {
				this.selectedDefDecoration = this.editor.deltaDecorations(this.selectedDefDecoration, []);
				return;
			}
			let range: monaco.IRange | null = null;
			for (let def of results) {
				if (def.range.startLineNumber === selection.startLineNumber && def.range.startColumn === selection.startColumn) {
					range = new monaco.Range(
						def.range.startLineNumber,
						def.range.startColumn,
						def.range.endLineNumber,
						def.range.endColumn,
					);
				}
			}
			if (!range) {
				return;
			}

			let decoration = {
				range: range,
				options: {
					inlineClassName: "selected-definition",
				},
			};
			this.selectedDefDecoration = this.editor.deltaDecorations(this.selectedDefDecoration, [decoration]);
		});
	}

	private addDecoration(range: monaco.IRange, text?: string): void {
		let model = this.editor.getModel();
		if (!model) {
			return;
		}
		this.decorations = this.editor.deltaDecorations(this.decorations, [{
			range: range,
			options: {
				inlineClassName: "goto-definition-link",
				hoverMessage: text,
			},
		}]);
	}

	private removeDecorations(): void {
		if (this.decorations.length > 0) {
			this.decorations = this.editor.deltaDecorations(this.decorations, []);
		}
	}

	private resetHandler(): void {
		this.lastMouseMoveEvent = null;
		this.removeDecorations();
	}

	private onEditorMouseUp(mouseEvent: monaco.editor.IEditorMouseEvent): void {
		if (!this.editor.getSelection().isEmpty()) {
			// Don't interfere with text selection.
			return;
		}

		if (mouseEvent.event.leftButton && mouseEvent.target.type === monaco.editor.MouseTargetType.CONTENT_TEXT && !mouseEvent.event.ctrlKey) {
			this.gotoDefinition(mouseEvent.target).done(() => {
				this.removeDecorations();
			}, (err: Error) => {
				this.removeDecorations();
				console.error(err);
			});
		}
	}

	private isEnabled(mouseEvent: monaco.editor.IEditorMouseEvent): boolean {
		// TODO(sqs): assumes that this is always true: DefinitionProviderRegistry.has(this.editor.getModel());
		return this.editor.getModel() &&
			(typeof mouseEvent.event.detail === "number" && mouseEvent.event.detail <= 1) &&
			mouseEvent.target.type === monaco.editor.MouseTargetType.CONTENT_TEXT;
	}

	private findDefinition(position: monaco.IPosition): Promise<monaco.languages.Location[]> {
		let model = this.editor.getModel();
		if (!model) {
			return Promise.resolve(null);
		}

		return new Promise((resolve, reject) => {
			(global as any).require(["vs/editor/contrib/goToDeclaration/common/goToDeclaration"], ({getDeclarationsAtPosition}) => {
				getDeclarationsAtPosition(model, position).then((result) => resolve(result));
			});
		});
	}

	private gotoDefinition(target: monaco.editor.IMouseTarget): monaco.Promise<any> {
		const model = this.editor.getModel();
		if (model) {
			const src = URI.repoParams(model.uri);
			EventLogger.logEventForCategory(
				AnalyticsConstants.CATEGORY_DEF,
				AnalyticsConstants.ACTION_CLICK,
				"BlobTokenClicked",
				{
					srcRepo: src.repo, srcRev: src.rev || "", srcPath: src.path,
					language: model.getModeId(),
				}
			);
		}

		// just run the corresponding action
		this.editor.setPosition(target.position);
		return this.editor.getAction("editor.action.goToDeclaration").run();
	}

	public getId(): string {
		return GotoDefinitionWithClickEditorContribution.ID;
	}

	public dispose(): void {
		this.decorations = this.editor.deltaDecorations(this.decorations, []);
		this.toUnhook.forEach(disposable => disposable.dispose());
	}
}

export interface ITask<T> {
	(): T;
}
