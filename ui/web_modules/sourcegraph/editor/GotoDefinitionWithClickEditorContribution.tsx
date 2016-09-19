// tslint:disable typedef ordered-imports
import {EventLogger} from "sourcegraph/util/EventLogger";
import {treeEntryFromUri} from "sourcegraph/editor/FileModel";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

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
	private currentWordUnderMouse: IWordAtPositionWithLine | null;
	private throttler: SimpleThrottler;
	private lastMouseMoveEvent: monaco.editor.IEditorMouseEvent | null;

	constructor(
		editor: monaco.editor.ICodeEditor,
	) {
		this.editor = editor;
		this.throttler = new SimpleThrottler();

		this.toUnhook.push(this.editor.onMouseUp((e: monaco.editor.IEditorMouseEvent) => this.onEditorMouseUp(e)));
		this.toUnhook.push(this.editor.onMouseMove((e: monaco.editor.IEditorMouseEvent) => this.onEditorMouseMove(e)));

		this.toUnhook.push(this.editor.onDidChangeCursorSelection((e) => this.onDidChangeCursorSelection(e)));
		this.toUnhook.push(this.editor.onDidChangeModel((e) => this.resetHandler()));
		this.toUnhook.push(this.editor.onDidChangeModelContent(() => this.resetHandler()));
		this.toUnhook.push(this.editor.onDidScrollChange((e) => {
			if (e.scrollTopChanged || e.scrollLeftChanged) {
				this.resetHandler();
			}
		}));
	}

	private onDidChangeCursorSelection(e: monaco.editor.ICursorSelectionChangedEvent): void {
		if (e.selection && e.selection.startColumn !== e.selection.endColumn) {
			this.resetHandler(); // immediately stop this feature if the user starts to select (https://github.com/Microsoft/vscode/issues/7827)
		}
	}

	private onEditorMouseMove(mouseEvent: monaco.editor.IEditorMouseEvent): void {
		if (mouseEvent.target.type === monaco.editor.MouseTargetType.UNKNOWN) {
			// Occurs occasionally when mousing over syntax-highlighted tokens. Must ignore or
			// else the decorations will erroneously be removed.
			return;
		}

		this.lastMouseMoveEvent = mouseEvent;
		this.startFindDefinition(mouseEvent);
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

		// Return early if word at position is still the same
		if (this.currentWordUnderMouse && this.currentWordUnderMouse.lineNumber === word.lineNumber && this.currentWordUnderMouse.startColumn === word.startColumn && this.currentWordUnderMouse.endColumn === word.endColumn && this.currentWordUnderMouse.word === word.word) {
			return;
		}

		this.currentWordUnderMouse = word;

		// Find definition and decorate word if found
		this.throttler.queue(() => this.findDefinition(mouseEvent.target)).then(results => {
			if (!results || !results.length) {
				this.removeDecorations();
				return;
			}

			this.addDecoration(
				{
					startLineNumber: position.lineNumber,
					startColumn: word.startColumn,
					endLineNumber: position.lineNumber,
					endColumn: word.endColumn,
				},
				results.length > 1 ? `Click to show the ${results.length} definitions found.` : undefined
			);
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
		if (mouseEvent.event.leftButton && mouseEvent.target.type === monaco.editor.MouseTargetType.CONTENT_TEXT) {
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

	private findDefinition(target: monaco.editor.IMouseTarget): Promise<monaco.languages.Location[]> {
		let model = this.editor.getModel();
		if (!model) {
			return Promise.resolve(null);
		}

		return new Promise((resolve, reject) => {
			(global as any).require(["vs/editor/contrib/goToDeclaration/common/goToDeclaration"], ({getDeclarationsAtPosition}) => getDeclarationsAtPosition(this.editor.getModel(), target.position).then((result) => resolve(result)));
		});
	}

	private gotoDefinition(target: monaco.editor.IMouseTarget): monaco.Promise<any> {
		const model = this.editor.getModel();
		if (model) {
			const src = treeEntryFromUri(model.uri);
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

export class SimpleThrottler {
	private current = Promise.resolve(null);

	queue<T>(promiseTask: ITask<Promise<T>>): Promise<T> {
		return this.current = this.current.then(() => promiseTask());
	}
}
