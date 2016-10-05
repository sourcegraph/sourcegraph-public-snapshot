// tslint:disable typedef ordered-imports
import {EventLogger} from "sourcegraph/util/EventLogger";
import {URIUtils} from "sourcegraph/core/uri";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import * as debounce from "lodash/debounce";

// tslint:disable typedef ordered-imports member-ordering
import {IEditorService} from "vs/platform/editor/common/editor";
import {editorContribution} from "vs/editor/browser/editorBrowserExtensions";
import * as editorCommon from "vs/editor/common/editorCommon";
import {ICodeEditor, IEditorMouseEvent, IMouseTarget} from "vs/editor/browser/editorBrowser";
import {IDisposable} from "vs/base/common/lifecycle";
import {Range} from "vs/editor/common/core/range";
import {TPromise} from "vs/base/common/winjs.base";
import {getDeclarationsAtPosition} from "vs/editor/contrib/goToDeclaration/common/goToDeclaration";
import {IKeyboardEvent} from "vs/base/browser/keyboardEvent";
import {Location} from "vs/editor/common/modes";
import {Selection} from "vs/editor/common/core/selection";
import {Position} from "vs/editor/common/core/position";

// IWordAtPositionWithLine lets us distinguish between two of the same
// words at the same start column on separate lines.
interface IWordAtPositionWithLine extends editorCommon.IWordAtPosition {
	lineNumber: number; // The line number where the word starts.
}

@editorContribution
export class GotoDefinitionWithClickEditorContribution implements editorCommon.IEditorContribution {
	private static ID = "editor.contrib.gotodefinitionwithclick";

	private toUnhook: IDisposable[] = [];
	private decorations: string[] = [];
	private selectedDefDecoration: string[] = [];
	private currentWordUnderMouse: IWordAtPositionWithLine | null;
	private lastMouseMoveEvent: IEditorMouseEvent | null;
	private findDefinitionDebounced: (target: IMouseTarget, word: editorCommon.IWordAtPosition) => void;
	private mouseLine: number;
	private mouseColumn: number;

	constructor(
		private editor: ICodeEditor,
		@IEditorService private editorService: IEditorService
	) {
		this.editor = editor;

		this.toUnhook.push(this.editor.onMouseUp((e: IEditorMouseEvent) => this.onEditorMouseUp(e)));
		this.toUnhook.push(this.editor.onMouseMove((e: IEditorMouseEvent) => this.onEditorMouseMove(e)));

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

	private onDidChangeCursorSelection(e: editorCommon.ICursorSelectionChangedEvent): void {
		if (e.selection && e.selection.startColumn !== e.selection.endColumn) {
			this.resetHandler(); // immediately stop this feature if the user starts to select (https://github.com/Microsoft/vscode/issues/7827)
		}

		// After the selection is changed check to see if the new current selection
		// has landed on a definition. If so, highlight it.
		this.highlightDefinitionAtSelection(e.selection);
	}

	private onEditorMouseMove(mouseEvent: IEditorMouseEvent): void {
		if (mouseEvent.target.type === editorCommon.MouseTargetType.UNKNOWN) {
			// Occurs occasionally when mousing over syntax-highlighted tokens. Must ignore or
			// else the decorations will erroneously be removed.
			return;
		}

		this.startFindDefinition(mouseEvent);
		this.lastMouseMoveEvent = mouseEvent;
	}

	private startFindDefinition(mouseEvent: IEditorMouseEvent, withKey?: IKeyboardEvent): void {
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

	private findDefinitionDebounced_(target: IMouseTarget, word: editorCommon.IWordAtPosition): void {
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

	private highlightDefinitionAtSelection(selection: Selection) {
		let position = ({
			lineNumber: selection.startLineNumber,
			column: selection.startColumn,
		});
		this.findDefinition(position).then(results => {
			if (!results || !results.length) {
				this.selectedDefDecoration = this.editor.deltaDecorations(this.selectedDefDecoration, []);
				return;
			}
			let range: editorCommon.IRange | null = null;
			for (let def of results) {
				if (def.range.startLineNumber === selection.startLineNumber && def.range.startColumn === selection.startColumn) {
					range = new Range(
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

	private addDecoration(range: editorCommon.IRange, text?: string): void {
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

	private onEditorMouseUp(mouseEvent: IEditorMouseEvent): void {
		if (!this.editor.getSelection().isEmpty()) {
			// Don't interfere with text selection.
			return;
		}

		if (mouseEvent.event.leftButton && mouseEvent.target.type === editorCommon.MouseTargetType.CONTENT_TEXT && !mouseEvent.event.ctrlKey) {
			this.gotoDefinition(mouseEvent.target).done(() => {
				this.removeDecorations();
			}, (err: Error) => {
				this.removeDecorations();
				console.error(err);
			});
		}
	}

	private isEnabled(mouseEvent: IEditorMouseEvent): boolean {
		// TODO(sqs): assumes that this is always true: DefinitionProviderRegistry.has(this.editor.getModel());
		return this.editor.getModel() &&
			(typeof mouseEvent.event.detail === "number" && mouseEvent.event.detail <= 1) &&
			mouseEvent.target.type === editorCommon.MouseTargetType.CONTENT_TEXT;
	}

	private findDefinition(position: editorCommon.IPosition): TPromise<Location[]> {
		let model = this.editor.getModel();
		if (!model) {
			throw new Error("no model");
		}

		return getDeclarationsAtPosition(model, Position.lift(position));
	}

	private gotoDefinition(target: IMouseTarget): TPromise<any> {
		const model = this.editor.getModel();
		if (model) {
			const src = URIUtils.repoParams(model.uri);
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
