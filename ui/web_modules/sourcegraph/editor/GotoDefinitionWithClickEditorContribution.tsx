import {URIUtils} from "sourcegraph/core/uri";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

import {IDisposable} from "vs/base/common/lifecycle";
import * as platform from "vs/base/common/platform";
import {TPromise} from "vs/base/common/winjs.base";
import {ICodeEditor, IEditorMouseEvent, IMouseTarget} from "vs/editor/browser/editorBrowser";
import {editorContribution} from "vs/editor/browser/editorBrowserExtensions";
import * as editorCommon from "vs/editor/common/editorCommon";
import {DefinitionProviderRegistry} from "vs/editor/common/modes";
import {IEditorService} from "vs/platform/editor/common/editor";

@editorContribution
export class GotoDefinitionWithClickEditorContribution implements editorCommon.IEditorContribution {
	private toUnhook: IDisposable[] = [];

	constructor(
		private editor: ICodeEditor,
		@IEditorService private editorService: IEditorService
	) {
		this.editor = editor;
		this.toUnhook.push(this.editor.onMouseUp((e: IEditorMouseEvent) => this.onEditorMouseUp(e)));
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
		// Don't use our contrib if they're pressing the goto def modifier already.
		const trigger = platform.isMacintosh ? "metaKey" : "ctrlKey";

		return this.editor.getModel() &&
			(typeof mouseEvent.event.detail === "number" && mouseEvent.event.detail <= 1) &&
			mouseEvent.target.type === editorCommon.MouseTargetType.CONTENT_TEXT &&
			!(platform.isMacintosh && mouseEvent.event.ctrlKey) &&
			!mouseEvent.event[trigger] &&
			DefinitionProviderRegistry.has(this.editor.getModel()) &&
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
		return "editor.contrib.gotodefinitionwithclick";
	}

	public dispose(): void {
		this.toUnhook.forEach(disposable => disposable.dispose());
	}
}
