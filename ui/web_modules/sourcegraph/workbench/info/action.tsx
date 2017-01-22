import { Subscription } from "rxjs-es";

import { IDisposable } from "vs/base/common/lifecycle";
import { ICodeEditor, IEditorMouseEvent } from "vs/editor/browser/editorBrowser";
import { editorContribution } from "vs/editor/browser/editorBrowserExtensions";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { IEditorContribution, IModel } from "vs/editor/common/editorCommon";

import { Features } from "sourcegraph/util/features";
import { fetchDependencyReferences, provideDefinition, provideGlobalReferences, provideReferences, provideReferencesCommitInfo } from "sourcegraph/util/RefsBackend";
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
	private globalFetchSubscription?: Promise<Subscription | undefined>;

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
			if (this.globalFetchSubscription) {
				this.globalFetchSubscription.then(sub => sub && sub.unsubscribe());
			}
			return;
		}
		const props = this.getParamsForMouseEvent(mouseEvent);
		if (!props) {
			return;
		}

		this.globalFetchSubscription = renderSidePanelForData(props);
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

export async function renderSidePanelForData(props: Props): Promise<Subscription | undefined> {
	const defDataP = provideDefinition(props.editorModel, props.lspParams.position);
	const referenceInfoP = provideReferences(props.editorModel, props.lspParams.position);
	const defData = await defDataP;
	if (!defData) {
		return;
	}
	infoStore.dispatch({ defData });

	const referenceInfo = await referenceInfoP;
	if (!referenceInfo || referenceInfo.length === 0) {
		infoStore.dispatch({ refModel: null, defData: defData });
		return;
	}

	let refModel = await provideReferencesCommitInfo(new ReferencesModel(referenceInfo, props.editorModel.uri));
	infoStore.dispatch({ defData, refModel });

	// Only fetch global refs for Go.
	if (props.editorModel.getModeId() !== "go") {
		return;
	}

	const depRefs = await fetchDependencyReferences(props.editorModel, props.lspParams.position);
	if (!depRefs) {
		return;
	}

	refModel = new ReferencesModel(referenceInfo, props.editorModel.uri);
	if (!refModel) {
		return;
	}

	infoStore.dispatch({ defData, refModel });

	let concatArray = referenceInfo;
	return provideGlobalReferences(props.editorModel, depRefs).subscribe(async refs => {
		concatArray = concatArray.concat(refs);

		refModel = new ReferencesModel(concatArray, props.editorModel.uri);

		if (!refModel) {
			return;
		}

		refModel = await provideReferencesCommitInfo(refModel);
		if (!refModel) {
			return;
		}

		infoStore.dispatch({ defData, refModel });
	});
}
