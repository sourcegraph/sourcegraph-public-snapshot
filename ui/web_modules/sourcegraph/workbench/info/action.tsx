import { IDisposable } from "vs/base/common/lifecycle";
import { ICodeEditor, IEditorMouseEvent } from "vs/editor/browser/editorBrowser";
import { editorContribution } from "vs/editor/browser/editorBrowserExtensions";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { IEditorContribution, IModel } from "vs/editor/common/editorCommon";

import { Features } from "sourcegraph/util/features";
import { DefinitionData, fetchDependencyReferencesReferences, provideDefinition, provideGlobalReferences, provideReferences } from "sourcegraph/util/RefsBackend";
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

const OpenInfoPanelID: string = "editor.contrib.openInfoPanel";
const TokenIdentifierClassName: string = "identifier";

@editorContribution
export class ReferenceAction implements IEditorContribution {
	private toDispose: IDisposable[] = [];

	public getId(): string {
		return OpenInfoPanelID;
	}

	public dispose(): void {
		this.toDispose.forEach(disposable => disposable.dispose());
	}

	constructor(
		private editor: ICodeEditor,
	) {
		if (editor instanceof EmbeddedCodeEditorWidget || !Features.projectWow.isEnabled()) {
			return;
		}
		this.toDispose.push(this.editor.onDidChangeModel((e) => {
			if (e.oldModelUrl !== null) {
				infoStore.dispatch(null);
			}
		}));
		this.toDispose.push(this.editor.onMouseUp((e: IEditorMouseEvent) => {
			this.onEditorMouseUp(e);
		}));
	}

	private async onEditorMouseUp(mouseEvent: IEditorMouseEvent): Promise<void> {
		if (!mouseEvent.event.target.classList.contains(TokenIdentifierClassName)) {
			// Hide side panel
			infoStore.dispatch(null);
			return;
		}
		const props = this.getParamsForMouseEvent(mouseEvent);
		if (!props) {
			return;
		}

		renderSidePanelForData(props);
	}

	private getParamsForMouseEvent(mouseEvent: IEditorMouseEvent): Props | null {
		if (!mouseEvent.event.target.classList.contains(TokenIdentifierClassName)) {
			return null;
		}

		const pos = mouseEvent.target.position;
		if (!pos) {
			return null;
		}
		const model = this.editor.getModel();

		const props: Props = {
			editorModel: model,
			lspParams: {
				position: {
					line: pos.lineNumber - 1,
					character: pos.column - 1,
				},
			}
		};

		return props;
	}
}

export async function renderSidePanelForData(props: Props): Promise<void> {
	const defData: DefinitionData | null = await provideDefinition(props.editorModel, props.lspParams.position);
	if (!defData) {
		return;
	}
	infoStore.dispatch({ defData });

	const referenceInfo = await provideReferences(props.editorModel, props.lspParams.position);
	if (!referenceInfo || referenceInfo.length === 0) {
		return;
	}

	let refModel = new ReferencesModel(referenceInfo);
	if (!refModel) {
		infoStore.dispatch({ refModel: null, defData: defData });
		return;
	}

	// refModel = await provideReferencesCommitInfo(refModel);

	infoStore.dispatch({ defData, refModel });

	const depRefs = await fetchDependencyReferencesReferences(props.editorModel, props.lspParams.position);
	if (!depRefs) {
		return;
	}

	const globalRefsModel = await provideGlobalReferences(depRefs);
	if (!globalRefsModel) {
		return;
	}

	let concatArray = globalRefsModel.concat(referenceInfo);
	refModel = new ReferencesModel(concatArray);

	if (!refModel) {
		return;
	}

	// Now go and fetch all the commit info for the new refs that we just got!
	// refModel = await provideReferencesCommitInfo(refModel);
	// if (!refModel) {
	// 	return;
	// }

	infoStore.dispatch({ defData, refModel });
}
