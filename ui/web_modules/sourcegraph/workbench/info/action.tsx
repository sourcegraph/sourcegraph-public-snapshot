import { IDisposable } from "vs/base/common/lifecycle";
import { ICodeEditor, IEditorMouseEvent } from "vs/editor/browser/editorBrowser";
import { editorContribution } from "vs/editor/browser/editorBrowserExtensions";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { IEditorContribution, IModel, IRange } from "vs/editor/common/editorCommon";

import { send } from "sourcegraph/editor/lsp";
import { Features } from "sourcegraph/util/features";
import { provideReferences, provideReferencesCommitInfo } from "sourcegraph/util/RefsBackend";
import { ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { infoStore } from "sourcegraph/workbench/info/sidebar";

export interface HoverData {
	definition: {
		uri: string;
		range: IRange;
	};
	docString: string;
	funcName: string;
}

interface Props {
	editorModel: IModel;
	lspParams: {
		position: {
			line: number,
			character: number
		},
		textDocument: { uri: string };
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

		// Load all data for the sidepane in chunks to prevent locking the UI on larger reference / commit fetches.
		const hoverData: HoverData | null = await this.fetchHoverData(props);
		if (!hoverData) {
			return;
		}
		infoStore.dispatch({ hoverData });

		let refModel = await resolveLocalReferences(props);

		if (!refModel) {
			infoStore.dispatch({ refModel: null, hoverData: hoverData });
			return;
		}

		refModel = await provideReferencesCommitInfo(refModel);

		infoStore.dispatch({ hoverData, refModel });
	}

	private async fetchHoverData(props: Props): Promise<HoverData | null> {
		const hoverInfo = await send(props.editorModel, "textDocument/hover", props.lspParams);
		if (!hoverInfo || !hoverInfo.result || !hoverInfo.result.contents) {
			return null;
		}
		const [{value: funcName}, docs] = hoverInfo.result.contents;
		const docString = docs ? docs.value : "";

		const defResponse = await send(props.editorModel, "textDocument/definition", props.lspParams);
		if (!defResponse.result || !defResponse.result[0]) {
			return null;
		}

		const defFirst = defResponse.result[0];
		let definition = {
			uri: defFirst.uri,
			range: {
				startLineNumber: defFirst.range.start.line,
				startColumn: defFirst.range.start.character,
				endLineNumber: defFirst.range.end.line,
				endColumn: defFirst.range.end.character,
			}
		};
		return { funcName, docString, definition };
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
				textDocument: {
					uri: model.uri.toString(true),
				}
			}
		};

		return props;
	}
}

async function resolveLocalReferences(props: Props): Promise<ReferencesModel | null> {
	const referenceInfo = await provideReferences(props.editorModel, props.lspParams.position);
	if (!referenceInfo || referenceInfo.length === 0) {
		return null;
	}
	return new ReferencesModel(referenceInfo);
}
