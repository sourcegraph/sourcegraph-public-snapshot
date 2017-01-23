import * as autobind from "autobind-decorator";
import { css } from "glamor";
import * as truncate from "lodash/truncate";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { RefTree } from "sourcegraph/workbench/info/refTree";
import { Location } from "vs/editor/common/modes";

import { FlexContainer, Heading, Loader } from "sourcegraph/components";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { DefinitionData } from "sourcegraph/util/RefsBackend";
import { DefinitionDocumentationHeader } from "sourcegraph/workbench/info/documentation";
import { Preview } from "sourcegraph/workbench/info/preview";
import { sidebarWidth } from "sourcegraph/workbench/info/preview";
import { ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
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
	private infoPanel: { open: boolean, id: string };
	private info: Props | null;

	constructor() {
		super();
		this.node = null;
		this.infoPanel = { open: false, id: "" };
		this.info = null;
	}

	componentWillMount(): void {
		this.toDispose.add(infoStore.subscribe((info) => {
			if (info && info.prepareData) {
				this.info = info;
				this.infoPanel = { open: info.prepareData.open, id: info.id };
				this.forceUpdate();
				return;
			}

			this.info = info;
			this.forceUpdate();
		}));
	}

	componentWillUnmount(): void {
		this.toDispose.dispose();
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

		// Optimistically assume it's the correct data.
		if (this.info && this.infoPanel.id === this.info.id && this.infoPanel.open) {
			ReactDOM.render(<InfoPanel {...this.info} />, node);
		} else {
			ReactDOM.render(<div></div>, node);
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
	id: string;
	defData: DefinitionData | null;
	refModel?: ReferencesModel | null;
	prepareData?: { open: boolean };
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

	private sidebarFunctionHeader(defData: DefinitionData | null): JSX.Element {
		let funcName = defData ? defData.funcName : "";
		return <FlexContainer justify="between" items="center" style={{
			backgroundColor: colors.blue(),
			boxShadow: `0 1px 2px 0 ${colors.black(0.3)}`,
			flex: "0 0 auto",
		}}>
			<Heading color="white" level={6} compact={true} style={{
				padding: whitespace[3],
				fontFamily: typography.fontStack.code,
				fontWeight: "normal",
				wordBreak: "break-word",
			}}>{truncate(funcName, { length: 120 })}</Heading>
			<div onClick={() => infoStore.dispatch({ defData: null, prepareData: { open: false }, id: "" })}
				{...css(
					{
						alignSelf: "flex-start",
						color: colors.blueGrayD1(),
						cursor: "pointer",
						padding: whitespace[3]
					},
					{ ":hover": { color: colors.blueGrayD2() } },
				) }>
				<Close width={18} />
			</div>
		</FlexContainer>;
	}

	render(): JSX.Element {
		const { defData, refModel } = this.props;
		const dividerSx = { width: "100%", borderColor: colors.blueGrayL2(0.3), margin: 0 };
		// position child elements relative to editor container
		(css as any).global(".editor-container", { position: "relative" });
		return <div>
			<FlexContainer direction="top_bottom" style={{
				position: "absolute",
				backgroundColor: "white",
				height: "100%",
				width: sidebarWidth,
				top: 0,
				right: 0,
				overflowY: "hidden",
			}}>
				{this.sidebarFunctionHeader(defData)}
				{defData && <DefinitionDocumentationHeader defData={defData} />}
				<hr style={dividerSx} />
				<div id={REFERENCES_SECTION_ID}>
					<FlexContainer items="center" style={{ height: 35, padding: whitespace[3] }}>
						<Heading level={7} color="gray" compact={true}>
							References
						</Heading>
					</FlexContainer>
				</div>
				<hr style={dividerSx} />
				{refModel === undefined && <div style={{ textAlign: "center" }}><Loader /></div>}
				{refModel === null && <div style={{ textAlign: "center" }}>No references</div>}
				{refModel && <RefTree
					model={refModel}
					focus={this.focusResource} />}
			</FlexContainer>
			<Preview
				location={this.state.previewLocation}
				hidePreview={this.refsFocused} />
		</div>;
	}
};

export const infoStore = new MiniStore<Props | null>();
