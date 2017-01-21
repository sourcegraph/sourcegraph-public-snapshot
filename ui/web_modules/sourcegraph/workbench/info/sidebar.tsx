import * as autobind from "autobind-decorator";
import * as truncate from "lodash/truncate";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { RefTree } from "sourcegraph/workbench/info/refTree";
import { Location } from "vs/editor/common/modes";
import { IModelService } from "vs/editor/common/services/modelService";

import { FlexContainer, Loader } from "sourcegraph/components";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { DefinitionData } from "sourcegraph/util/RefsBackend";
import { renderSidePanelForData } from "sourcegraph/workbench/info/action";
import { DefinitionDocumentationHeader } from "sourcegraph/workbench/info/documentation";
import { Preview } from "sourcegraph/workbench/info/preview";
import { sidebarWidth } from "sourcegraph/workbench/info/preview";
import { ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { Services } from "sourcegraph/workbench/services";
import { Disposables } from "sourcegraph/workbench/utils";
import { MiniStore } from "sourcegraph/workbench/utils";

export const REFERENCES_SECTION_ID: string = "ReferencesSectionID";
const TreeDomNodeID: string = "workbench.editors.stringEditor";
const TreeSidebarClassName: string = "sg-sidebar";

export interface InfoPanelProps {
	isSymbolUrl: boolean;
	repo: GQL.IRepository;
}

// Lifecycle methods for the InfoPanel. Doesn't render a node in the tree.
@autobind
export class InfoPanelLifecycle extends React.Component<InfoPanelProps, {}> {
	private toDispose: Disposables = new Disposables();
	private node: HTMLDivElement | null;
	private infoPaneExpanded: boolean;
	private info: Props | null;

	constructor() {
		super();
		this.node = null;
		this.infoPaneExpanded = false;
		this.info = null;
	}

	componentDidMount(): void {
		this.openSideBarForSymbolURL(this.props);
	}

	componentWillMount(): void {
		this.toDispose.add(infoStore.subscribe((info) => {
			this.info = info;
			this.forceUpdate();
		}));
	}

	componentWillUnmount(): void {
		this.toDispose.dispose();
	}

	private openSideBarForSymbolURL(props: InfoPanelProps): void {
		if (this.props.isSymbolUrl && this.props.repo && this.props.repo.symbols && this.props.repo.symbols.length > 0) {
			const {line, character, path} = this.props.repo.symbols[0];
			const lspParams = {
				position: {
					line,
					character
				},
			};

			this.toDispose.add(Services.get<IModelService>(IModelService).onModelAdded((editorModel) => {
				if (`${editorModel.uri.authority}${editorModel.uri.path}` === this.props.repo.uri) {
					if (editorModel.uri.fragment === path) {
						renderSidePanelForData({ editorModel, lspParams });
					}
				}
			}));
		}
	}

	renderOutOfTreeDOMNode(): void {
		const parent = document.getElementById(TreeDomNodeID) as HTMLDivElement;
		if (!parent) {
			return;
		}

		const children = parent.getElementsByClassName(TreeSidebarClassName);
		for (let i = 0; i < children.length; i++) {
			parent.removeChild(children[i]);
		}

		const node = document.createElement("div");
		node.className = TreeSidebarClassName;
		parent.appendChild(node);

		if (this.info === null) {
			ReactDOM.render(<div></div>, node);
		} else {
			ReactDOM.render(<InfoPanel {...this.info} />, node);
		}
	}

	render(): null {
		setTimeout(() => {
			this.renderOutOfTreeDOMNode();
		});
		return null;
	}
}

interface State {
	previewLocation: Location | null;
}

export interface Props {
	defData: DefinitionData;
	refModel?: ReferencesModel | null;
};

const style = {
	position: "absolute",
	top: layout.editorToolbarHeight,
	background: "white",
	height: `calc(100% - ${layout.editorToolbarHeight}px)`,
	width: sidebarWidth,
	right: "0px",
	overflowY: "hidden",
	display: "flex",
	flexDirection: "column",
	color: "black",
};

@autobind
class InfoPanel extends React.Component<Props, State> {

	constructor() {
		super();
		this.state = {
			previewLocation: null,
		};
	}

	private refsFocused(): void {
		this.setState({
			previewLocation: null,
		});
	}

	private focusResource(loc: Location): void {
		this.setState({
			previewLocation: loc,
		});
	}

	private sidebarFunctionHeader(defData: DefinitionData): JSX.Element {
		return (
			<div>
				<FlexContainer style={{ backgroundColor: "#1893e7", boxShadow: "0 1px 2px 0 rgba(0, 0, 0, 0.16)" }}>
					<code style={Object.assign({ color: "white", paddingLeft: whitespace[3], paddingTop: whitespace[3], paddingBottom: whitespace[3], paddingRight: whitespace[2] }, typography[3])}>{truncate(defData.funcName, { length: 120 })}</code>
					<span onClick={() => infoStore.dispatch(null)} style={{ cursor: "pointer", marginLeft: "auto", paddingRight: whitespace[3], paddingTop: whitespace[3], }}>
						<Close width={18} color={colors.blueGrayD1(0.5)} />
					</span>
				</FlexContainer>
			</div>
		);
	}

	render(): JSX.Element {
		const { defData, refModel } = this.props;
		return <div>
			<div style={style}>
				{this.sidebarFunctionHeader(defData)}
				<DefinitionDocumentationHeader
					defData={defData} />
				<div style={{ paddingTop: whitespace[2] }}>
					<div style={{ width: "100%", height: "1px", backgroundColor: "rgba(201, 211, 227, 0.3)" }} />
				</div>
				<div id={REFERENCES_SECTION_ID}>
					<FlexContainer items="center" style={{ height: 35, padding: whitespace[2] }}>
						<span style={Object.assign({},
							{
								marginTop: "10px",
								fontSize: "11px",
								fontWeight: "bold",
								fontStyle: "normal",
								fontStretch: "normal",
								color: colors.blueGray(),
							},
							typography[1])}>
							REFERENCES:
					</span>
					</FlexContainer>
				</div>
				<div style={{ paddingTop: whitespace[2] }}>
					<div style={{ width: "100%", height: "1px", backgroundColor: "rgba(201, 211, 227, 0.3)" }} />
				</div>
				{refModel === undefined && <div style={{ textAlign: "center" }}><Loader /></div>}
				{refModel === null && <div style={{ textAlign: "center" }}>No references</div>}
				{refModel && <RefTree
					model={refModel}
					focus={this.focusResource} />}
			</div>
			<Preview
				location={this.state.previewLocation}
				hidePreview={this.refsFocused} />
		</div>;
	}
}

export const infoStore = new MiniStore<Props | null>();
