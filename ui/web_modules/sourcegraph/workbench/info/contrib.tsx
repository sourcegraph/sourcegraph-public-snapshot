import { Subscription } from "rxjs";

import { Events, FileEventProps } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { KeyCode, KeyMod } from "vs/base/common/keyCodes";
import { isMacintosh } from "vs/base/common/platform";
import { ICodeEditor, IEditorMouseEvent } from "vs/editor/browser/editorBrowser";
import { editorContribution } from "vs/editor/browser/editorBrowserExtensions";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { IPosition } from "vs/editor/common/editorCommon";
import { IModel } from "vs/editor/common/editorCommon";
import { CommonEditorRegistry } from "vs/editor/common/editorCommonExtensions";
import { getOuterEditor } from "vs/editor/contrib/zoneWidget/browser/peekViewWidget";
import { ContextKeyExpr } from "vs/platform/contextkey/common/contextkey";
import { IEditorService } from "vs/platform/editor/common/editor";
import { KeybindingsRegistry } from "vs/platform/keybinding/common/keybindingsRegistry";

import { URIUtils } from "sourcegraph/core/uri";
import { normalisePosition } from "sourcegraph/editor/contrib";
import { DefinitionData, fetchDependencyReferences, provideDefinition, provideGlobalReferences, provideReferences, provideReferencesCommitInfo } from "sourcegraph/util/RefsBackend";
import { ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { infoStore } from "sourcegraph/workbench/info/sidebar";
import { Services } from "sourcegraph/workbench/services";
import { Disposables } from "sourcegraph/workbench/utils";

interface Props {
	editorModel: IModel;
	lspParams: {
		position: {
			line: number,
			character: number
		},
	};
};

export const SidebarContribID = "sg.contrib.openSidePanel";

@editorContribution
export class SidebarContribution extends Disposables {

	private globalFetchSubscription?: Promise<Subscription | undefined>;
	currentID: string = "";

	constructor(
		private editor: ICodeEditor,
	) {
		super();

		if (editor instanceof EmbeddedCodeEditorWidget) {
			this.add(this.editor.onMouseUp(this.peekViewMouseUp));
			return;
		}

		this.add(this.editor.onMouseUp(this.mouseUp));

		editor.onDidChangeModel(event => {
			let oldModel = event.oldModelUrl;
			let newModel = event.newModelUrl;
			if (!oldModel || (newModel && oldModel.toString() !== newModel.toString())) {
				const fileEventProps = URIUtils.repoParams(newModel);
				this.prepareInfoStore(false, "", fileEventProps);
			}
		});
	}

	getId(): string {
		return SidebarContribID;
	}

	private async renderSidePanelForData(props: Props): Promise<Subscription | undefined> {
		const id = this.currentID;
		const fileEventProps = URIUtils.repoParams(props.editorModel.uri);
		const def: DefinitionData | null = await provideDefinition(props.editorModel, props.lspParams.position).then(defData => {
			if (!defData || (!defData.docString && !defData.funcName)) {
				return null;
			}
			this.prepareInfoStore(true, this.currentID, fileEventProps);
			this.dispatchInfo(id, defData, fileEventProps);
			return defData;
		});

		const localRefs = await provideReferences(props.editorModel, props.lspParams.position);
		if (this.currentID !== id) {
			return;
		}

		if (!localRefs || localRefs.length === 0) {
			this.dispatchInfo(id, def, fileEventProps, null);
		}

		let refModel = new ReferencesModel(localRefs, props.editorModel.uri);
		this.dispatchInfo(id, def, fileEventProps, refModel);

		const localRefsWithCommitInfo = await provideReferencesCommitInfo(localRefs);
		refModel = new ReferencesModel(localRefsWithCommitInfo, props.editorModel.uri);
		if (this.currentID !== id) {
			return;
		}
		this.dispatchInfo(id, def, fileEventProps, refModel);

		const depRefs = await fetchDependencyReferences(props.editorModel, props.lspParams.position);
		if (!depRefs || this.currentID !== id) {
			this.dispatchInfo(id, def, fileEventProps, refModel, true);
			return;
		}

		let localAndGlobalRefs = localRefsWithCommitInfo;
		return provideGlobalReferences(props.editorModel, depRefs).subscribe(refs => {
			localAndGlobalRefs = localAndGlobalRefs.concat(refs);
			refModel = new ReferencesModel(localAndGlobalRefs, props.editorModel.uri);
			this.dispatchInfo(id, def, fileEventProps, refModel);
		}, () => null, () => this.dispatchInfo(id, def, fileEventProps, refModel, true));
	}

	private prepareInfoStore(prepare: boolean, id: string, fileEventProps: FileEventProps): void {
		if (!prepare && this.globalFetchSubscription) {
			this.globalFetchSubscription.then(sub => sub && sub.unsubscribe());
		}
		infoStore.dispatch({ defData: null, prepareData: { open: prepare }, loadingComplete: true, id, fileEventProps });
	}

	private dispatchInfo(id: string, defData: DefinitionData | null, fileEventProps: FileEventProps, refModel?: ReferencesModel | null, loadingComplete?: boolean): void {
		if (id === this.currentID) {
			infoStore.dispatch({ defData, refModel, loadingComplete, id, fileEventProps });
		} else if (this.globalFetchSubscription) {
			this.globalFetchSubscription.then(sub => sub && sub.unsubscribe());
		}
	}

	private shouldTrigger(e: IEditorMouseEvent): boolean {
		if (e.event.ctrlKey) {
			return false;
		}
		if (e.event.metaKey && isMacintosh) {
			return false;
		}
		const sel = this.editor.getSelection();
		if (!sel.isEmpty()) {
			return false;
		}
		return true;
	}

	/**
	 * When we click on a token in the peek view, we should close the peek
	 * view, open that file in the main view, and open the side panel for that
	 * token.
	 */
	private peekViewMouseUp = (e: IEditorMouseEvent): void => {
		this.logClick(e);
		if (!this.shouldTrigger(e)) {
			return;
		}
		const editorService = Services.get(IEditorService) as IEditorService;
		const model = this.editor.getModel();
		const pos = normalisePosition(model, this.editor.getPosition());
		const selection = this.editor.getSelection();
		if (this.isIdentifier(model, pos)) {
			editorService.openEditor({ resource: model.uri, options: { selection } });
		}
	}

	private mouseUp = (e: IEditorMouseEvent): void => {
		this.logClick(e);
		if (!this.shouldTrigger(e)) {
			return;
		}

		this.openInSidebar();
	}

	public openInSidebar = (): void => {
		const model = this.editor.getModel();
		const pos = normalisePosition(model, this.editor.getPosition());
		this.currentID = model.uri.toString() + pos.lineNumber + ":" + pos.column;

		if (!this.isIdentifier(model, pos)) {
			return;
		}

		this.highlightWord(pos);

		this.globalFetchSubscription = this.renderSidePanelForData({
			editorModel: model,
			lspParams: {
				position: {
					line: pos.lineNumber - 1,
					character: pos.column - 1,
				},
			},
		});
	}

	private highlightWord(position: IPosition): void {
		const model = this.editor.getModel();

		const word = model.getWordAtPosition(position);
		if (!word) {
			return;
		}

		this.editor.setSelection({
			startLineNumber: position.lineNumber,
			startColumn: word.startColumn,
			endLineNumber: position.lineNumber,
			endColumn: word.endColumn,
		});
	}

	private isIdentifier(model: IModel, pos: IPosition): boolean {
		const line = model.getLineTokens(pos.lineNumber);
		const tokens = line.sliceAndInflate(pos.column, pos.column, 0);
		if (tokens.length !== 1) {
			return true;
		}
		const token = tokens[0];
		if (token.type.length === 0) {
			// Model hasn't been tokenized yet.
			return true;
		}
		return token.type.includes("identifier");
	}

	private logClick(e: IEditorMouseEvent): void {
		const model = this.editor.getModel();

		const params = URIUtils.repoParams(model.uri);
		if (this.editor instanceof EmbeddedCodeEditorWidget) {
			const resource = getOuterEditor(Services, {}).getModel().uri;
			const outerParams = URIUtils.repoParams(resource);
			Events.CodeToken_Clicked.logEvent({
				...outerParams,
				refRepo: params.repo,
				refRev: params.rev,
				refPath: params.path,
				language: model.getModeId(),
			});
			return;
		}

		Events.CodeToken_Clicked.logEvent({
			...params,
			language: model.getModeId(),
		});
	}

}

function closeActiveReferenceSearch(): void {
	infoStore.dispatch({ defData: null, prepareData: { open: false }, loadingComplete: true, id: "", fileEventProps: { repo: "", rev: null, path: "" } });
}

KeybindingsRegistry.registerCommandAndKeybindingRule({
	id: "closeSidePaneWidget",
	weight: CommonEditorRegistry.commandWeight(-202),
	primary: KeyCode.Escape,
	secondary: [KeyMod.Shift | KeyCode.Escape | KeyCode.Delete], // tslint:disable-line
	when: ContextKeyExpr.and(ContextKeyExpr.not("config.editor.stablePeek")),
	handler: closeActiveReferenceSearch
});
