import {urlToBlobLineCol} from "sourcegraph/blob/routes";
import {URIUtils} from "sourcegraph/core/uri";
import * as lsp from "sourcegraph/editor/lsp";
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

		if (this.newTabEvent(mouseEvent)) {
			// HACK: Disable the builtin gotodef contrib for this event.
			mouseEvent.event.metaKey = false;
			mouseEvent.event.ctrlKey = false;

			this.openInNewTab(mouseEvent.target);
			return;
		}

		if (this.isEnabled(mouseEvent)) {
			this.gotoDefinition(mouseEvent.target);
		}
	}

	private isEnabled(mouseEvent: IEditorMouseEvent): boolean {
		return this.editor.getModel() &&
			(typeof mouseEvent.event.detail === "number" && mouseEvent.event.detail <= 1) &&
			mouseEvent.target.type === editorCommon.MouseTargetType.CONTENT_TEXT &&
			!(platform.isMacintosh && mouseEvent.event.ctrlKey) &&
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

	private newTabEvent(mouseEvent: IEditorMouseEvent): boolean {
		const ctrl = platform.isMacintosh ? "metaKey" : "ctrlKey";
		return mouseEvent.event.middleButton || (mouseEvent.event.leftButton && mouseEvent.event[ctrl]);
	}

	// openInNewTab opens the definition in a new tab.
	private async openInNewTab(target: IMouseTarget): Promise<void> {
		const model = this.editor.getModel();
		if (model === null) {
			return;
		}
		const {repo, rev, path} = URIUtils.repoParams(model.uri);
		AnalyticsConstants.Events.CodeToken_Clicked.logEvent({
			srcRepo: repo, srcRev: rev || "", srcPath: path,
			language: model.getModeId(),
		});
		const params = {
			position: lsp.toPosition(target.position),
			textDocument: {
				uri: model.uri.toString(true),
			},
		};
		// We have to create the tab before the async call to prevent it from
		// being blocked as a popup.
		const tab = global.window.open("", "_blank");

		try {
			const response = await lsp.send(model, "textDocument/definition", params);
			const location = response && checkResponse(response.result);
			if (!location) {
				// If we didn't click on something useful, close the tab.
				tab.close();
			} else {
				tab.location = urlForLocation(location);
			}
		} catch (err) {
			// There are a lot of places this can throw, so make sure we clean
			// up.
			tab.close();
		}
	}

}

// checkResponse checks if we got a result, and that it conforms to the
// lsp.Location interface.
function checkResponse(result: any): lsp.Location | null {
	if (result.length > 0) {
		result = result[0];
	}
	if (result["uri"] && result["range"]) {
		return result;
	}
	return null;
}

function urlForLocation(loc: lsp.Location): string {
	const {line, character} = loc.range.start;
	const {repo, rev, path} = URIUtils.repoParamsExt(loc.uri);
	return urlToBlobLineCol(repo, rev, path, line + 1, character + 1);
}
