import { Subscription } from "rxjs";

import { FileEventProps } from "sourcegraph/util/constants/AnalyticsConstants";
import { KeyCode, KeyMod } from "vs/base/common/keyCodes";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import * as editorCommon from "vs/editor/common/editorCommon";
import { IModel } from "vs/editor/common/editorCommon";
import { EditorAction, ServicesAccessor, editorAction } from "vs/editor/common/editorCommonExtensions";
import { CommonEditorRegistry } from "vs/editor/common/editorCommonExtensions";
import { PeekContext } from "vs/editor/contrib/zoneWidget/browser/peekViewWidget";
import { ContextKeyExpr } from "vs/platform/contextkey/common/contextkey";
import { IEditorService } from "vs/platform/editor/common/editor";
import { KeybindingsRegistry } from "vs/platform/keybinding/common/keybindingsRegistry";

import { URIUtils } from "sourcegraph/core/uri";
import { normalisePosition } from "sourcegraph/editor/contrib";
import { DefinitionData, fetchDependencyReferences, provideDefinition, provideGlobalReferences, provideReferences, provideReferencesCommitInfo } from "sourcegraph/util/RefsBackend";
import { ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { infoStore } from "sourcegraph/workbench/info/sidebar";

interface Props {
	editorModel: IModel;
	lspParams: {
		position: {
			line: number,
			character: number
		},
	};
};

export const SidebarActionID = "editor.action.openSidePanel";

@editorAction
export class DefinitionAction extends EditorAction {

	private globalFetchSubscription?: Promise<Subscription | undefined>;
	currentID: string = "";

	constructor() {
		super({
			id: SidebarActionID,
			label: "Open Side Panel",
			alias: "Open Side Panel",
			precondition: editorCommon.ModeContextKeys.hasDefinitionProvider,
		});
	}

	public run(accessor: ServicesAccessor, editor: editorCommon.ICommonCodeEditor): void {
		const editorService = accessor.get(IEditorService);

		editor.onDidChangeModel(event => {
			let oldModel = event.oldModelUrl;
			let newModel = event.newModelUrl;
			if (!oldModel || (newModel && oldModel.toString() !== newModel.toString())) {
				const eventProps = URIUtils.repoParams(newModel);
				this.prepareInfoStore(false, "", eventProps);
			}
		});

		this.onResult(editorService, editor);
	}

	async renderSidePanelForData(props: Props): Promise<Subscription | undefined> {
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

	private onResult(editorService: IEditorService, editor: editorCommon.ICommonCodeEditor): void {
		const model = editor.getModel();
		const eventProps = URIUtils.repoParams(model.uri);
		const position = normalisePosition(model, editor.getPosition());
		const word = model.getWordAtPosition(position);
		if (!word) {
			return;
		}

		this.currentID = model.uri.toString() + position.lineNumber + ":" + position.column;

		if (editor instanceof EmbeddedCodeEditorWidget) {
			const range = {
				startLineNumber: position.lineNumber,
				startColumn: word.startColumn,
				endLineNumber: position.lineNumber,
				endColumn: word.endColumn,
			};
			this.prepareInfoStore(true, this.currentID, eventProps);
			editorService.openEditor({
				resource: model.uri,
				options: {
					selection: range,
					revealIfVisible: true,
				}
			}, true).then(() => {
				this.openInSidebar(editor, position);
			});

			return;
		}

		if (ContextKeyExpr.and(editorCommon.ModeContextKeys.hasDefinitionProvider, PeekContext.notInPeekEditor)) {
			this.openInSidebar(editor, position);
		}
	}

	private openInSidebar(editor: editorCommon.ICommonCodeEditor, pos: editorCommon.IPosition): void {
		this.globalFetchSubscription = this.renderSidePanelForData({
			editorModel: editor.getModel(),
			lspParams: {
				position: {
					line: pos.lineNumber - 1,
					character: pos.column - 1,
				},
			},
		});

		this.highlightWord(editor, pos);
	}

	private highlightWord(editor: editorCommon.ICommonCodeEditor, position: editorCommon.IPosition): void {
		const model = editor.getModel();
		if (!this.isIdentifier(model, position)) {
			return;
		}
		const word = model.getWordAtPosition(position);
		editor.setSelection({
			startLineNumber: position.lineNumber,
			startColumn: word.startColumn,
			endLineNumber: position.lineNumber,
			endColumn: word.endColumn,
		});
	}

	private isIdentifier(model: IModel, pos: editorCommon.IPosition): boolean {
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
