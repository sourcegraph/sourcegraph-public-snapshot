// tslint:disable typedef ordered-imports
import {URIUtils} from "sourcegraph/core/uri";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

// tslint:disable typedef ordered-imports member-ordering
import {IEditorService} from "vs/platform/editor/common/editor";
import {editorContribution} from "vs/editor/browser/editorBrowserExtensions";
import * as editorCommon from "vs/editor/common/editorCommon";
import {ICodeEditor, IEditorMouseEvent, IMouseTarget} from "vs/editor/browser/editorBrowser";
import {IDisposable} from "vs/base/common/lifecycle";
import {TPromise} from "vs/base/common/winjs.base";

// IWordAtPositionWithLine lets us distinguish between two of the same
// words at the same start column on separate lines.
interface IWordAtPositionWithLine extends editorCommon.IWordAtPosition {
	lineNumber: number; // The line number where the word starts.
}

@editorContribution
export class GotoDefinitionWithClickEditorContribution implements editorCommon.IEditorContribution {
	private static ID = "editor.contrib.gotodefinitionwithclick";

	private toUnhook: IDisposable[] = [];
	private mouseLine: number;
	private mouseColumn: number;

	constructor(
		private editor: ICodeEditor,
		@IEditorService private editorService: IEditorService
	) {
		this.editor = editor;

		this.toUnhook.push(this.editor.onMouseUp((e: IEditorMouseEvent) => this.onEditorMouseUp(e)));
		this.toUnhook.push(this.editor.onMouseMove((e) => {
			if (!e.target.position) {
				return;
			}
			this.mouseLine = e.target.position.lineNumber;
			this.mouseColumn = e.target.position.column;
		}));
	}

	private onEditorMouseUp(mouseEvent: IEditorMouseEvent): void {
		if (!this.editor.getSelection().isEmpty()) {
			// Don't interfere with text selection.
			return;
		}

		if (this.isEnabled(mouseEvent)) {
			this.gotoDefinition(mouseEvent.target);
		}
	}

	private isEnabled(mouseEvent: IEditorMouseEvent): boolean {
		// TODO(sqs): assumes that this is always true: DefinitionProviderRegistry.has(this.editor.getModel());
		return this.editor.getModel() &&
			(typeof mouseEvent.event.detail === "number" && mouseEvent.event.detail <= 1) &&
			mouseEvent.target.type === editorCommon.MouseTargetType.CONTENT_TEXT &&
			!mouseEvent.event.ctrlKey &&
			mouseEvent.event.leftButton;
	}

	private gotoDefinition(target: IMouseTarget): TPromise<any> {
		const model = this.editor.getModel();
		if (model) {
			const src = URIUtils.repoParams(model.uri);
			AnalyticsConstants.Events.CodeToken_Clicked.logEvent({
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
		this.toUnhook.forEach(disposable => disposable.dispose());
	}
}

export interface ITask<T> {
	(): T;
}
