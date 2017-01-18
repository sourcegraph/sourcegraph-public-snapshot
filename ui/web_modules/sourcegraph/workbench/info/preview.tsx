import * as autobind from "autobind-decorator";
import * as React from "react";
import { Link } from "react-router";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { IEditorOptions } from "vs/editor/common/editorCommon";
import { Location } from "vs/editor/common/modes";

import { urlToBlobLine } from "sourcegraph/blob/routes";
import { colors } from "sourcegraph/components/utils";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance } from "sourcegraph/editor/Editor";
import { REFERENCES_SECTION_ID, infoStore } from "sourcegraph/workbench/info/sidebar";
import { Services } from "sourcegraph/workbench/services";
import { Disposables, RouterContext, scrollToLine } from "sourcegraph/workbench/utils";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IEditorService } from "vs/platform/editor/common/editor";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";

interface Props {
	location: Location | null;
	hidePreview: () => void;
}

export const funcHeight = 60;
export const sidebarWidth = 300;
const titleHeight = 30;

export class Preview extends React.Component<Props, {}> {
	render(): JSX.Element {
		if (!this.props.location) {
			return <div></div>;
		}
		let element = document.getElementById(REFERENCES_SECTION_ID);
		let boundingRect;
		if (element) {
			boundingRect = element.getBoundingClientRect();
		}
		let top = (boundingRect as any).top - 40;
		return <div style={{
			height: "100%",
			position: "absolute",
			width: "100%",
			bottom: "0px",
		}}>
			<div style={{
				height: `${top}px`,
				background: "rgba(20, 20, 20, 0.46)",
				zIndex: 1,
			}}
				onClick={this.props.hidePreview}
				></div>
			<div style={{
				width: `calc(100% - ${sidebarWidth}px)`,
				height: `calc(100% - ${top}px)`,
				position: "absolute",
				bottom: "0px",
			}}>
				<Title location={this.props.location} />
				<EditorPreview location={this.props.location} />
			</div>
		</div>;
	}
}

const prefix = "github.com/";

function Title(props: EditorProps): JSX.Element {
	let {repo, path, rev} = URIUtils.repoParams(props.location.uri);
	const url = urlToBlobLine(repo, rev, path, props.location.range.startLineNumber);
	repo = repo.startsWith(prefix) ? repo.substr(prefix.length) : repo;
	return <div style={{
		background: colors.blue(),
		color: colors.white(),
		height: titleHeight,
		lineHeight: `${titleHeight}px`,
		paddingLeft: 20,
	}}>
		<RouterContext>
			<Link
				to={url}
				style={{
					color: colors.white()
				}}>
				{repo}/{path}
			</Link>
		</RouterContext>
	</div>;
}

interface EditorProps {
	location: Location;
}

@autobind
class EditorPreview extends React.Component<EditorProps, {}> {

	private toDispose: Disposables = new Disposables();
	private preview: EmbeddedCodeEditorWidget;

	componentWillUnmount(): void {
		this.toDispose.dispose();
	}

	private editorDiv(div: HTMLDivElement): void {
		if (div === null) {
			return;
		}

		const instantiationService = Services.get(IInstantiationService);
		const editor = getEditorInstance();

		const options: IEditorOptions = {
			scrollBeyondLastLine: false,
			overviewRulerLanes: 2,
			fixedOverflowWidgets: true
		};

		this.preview = instantiationService.createInstance(EmbeddedCodeEditorWidget, div, options, editor);
		this.preview.onMouseUp(this.selectReference);
		this.preview.layout();
		this.toDispose.add(this.preview);
		this.setContents();
	}

	private selectReference(): void {
		const editorService = Services.get(IEditorService);
		editorService.openEditor({ resource: this.props.location.uri }).then(() => {
			const editor = getEditorInstance();
			scrollToLine(editor, this.props.location.range.startLineNumber);
			infoStore.dispatch(null);
		});
	}

	private setContents(): void {
		if (!this.preview) {
			return;
		}
		const modelService = Services.get(ITextModelResolverService);
		modelService.createModelReference(this.props.location.uri).then((ref) => {
			const model = ref.object.textEditorModel;
			this.preview.layout();
			this.preview.setModel(model);
			this.preview.setSelection(this.props.location.range, true, true);
		});
	}

	render(): JSX.Element {
		this.setContents();
		return <div style={{
			height: `calc(100% - ${titleHeight}px)`
		}} ref={this.editorDiv}>
		</div>;
	}
}
