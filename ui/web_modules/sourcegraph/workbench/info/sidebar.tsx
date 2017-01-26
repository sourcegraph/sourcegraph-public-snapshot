import * as autobind from "autobind-decorator";
import { css } from "glamor";
import * as truncate from "lodash/truncate";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { RefTree } from "sourcegraph/workbench/info/refTree";
import { Location } from "vs/editor/common/modes";

import { FlexContainer, Heading, Panel } from "sourcegraph/components";
import { Spinner } from "sourcegraph/components/symbols";
import { Close, Report } from "sourcegraph/components/symbols/Primaries";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { DefinitionData } from "sourcegraph/util/RefsBackend";
import { DefinitionDocumentationHeader } from "sourcegraph/workbench/info/documentation";
import { Preview } from "sourcegraph/workbench/info/preview";
import { sidebarWidth } from "sourcegraph/workbench/info/preview";
import { ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { Disposables } from "sourcegraph/workbench/utils";
import { MiniStore } from "sourcegraph/workbench/utils";

export const REFERENCES_SECTION_ID: string = "references-section-header";
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
	private infoPanelRef: InfoPanel | HTMLElement;
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
				// Close preview when escape key is clicked - Don't dismiss sidepane.
				if (!info.prepareData.open && !info.id && this.infoPanelRef instanceof InfoPanel && this.infoPanelRef.state.previewLocation) {
					this.infoPanelRef.setState({
						previewLocation: null,
						refModel: this.infoPanelRef.state.refModel,
					});
					return;
				}
				this.info = info;
				this.infoPanel = { open: info.prepareData.open, id: info.id };
				this.forceUpdate();
				return;
			}
			if (info && info.loadingComplete !== undefined && this.infoPanelRef instanceof InfoPanel) {
				this.infoPanelRef.setState({
					loadingComplete: info.loadingComplete,
				});

				return;
			}
			if (info && info.refModel !== undefined && this.infoPanelRef && this.infoPanelRef instanceof InfoPanel) {
				this.info = info;
				const currentSelected = this.infoPanelRef.state.previewLocation;
				this.infoPanelRef.setState({
					previewLocation: currentSelected,
					refModel: info.refModel,
				});

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
			ReactDOM.render(<InfoPanel ref={(el) => this.infoPanelRef = el} {...this.info} />, node);
		} else {
			ReactDOM.render(<div ref={(el) => this.infoPanelRef = el}></div>, node);
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
	previewLocation?: Location | null;
	refModel?: ReferencesModel | null;
	loadingComplete?: boolean;
}

export interface Props {
	id: string;
	defData: DefinitionData | null;
	refModel?: ReferencesModel | null;
	prepareData?: { open: boolean };
	loadingComplete?: boolean;
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
				overflowWrap: "break-word",
				overflow: "hidden",
			}}>{truncate(funcName, { length: 120 })}</Heading>
			<div onClick={() => infoStore.dispatch({ defData: null, prepareData: { open: false }, loadingComplete: true, id: this.props.id })}
				{...css(
					{
						alignSelf: "flex-start",
						color: colors.black(0.8),
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
		const { defData } = this.props;
		const { refModel, loadingComplete } = this.state;
		const dividerSx = { width: "100%", borderColor: colors.blueGrayL2(0.3), margin: 0 };
		const refsLoading = refModel !== null && !loadingComplete;
		// position child elements relative to editor container
		return <div style={{ height: "100%" }}>
			<FlexContainer direction="top_bottom" style={{
				position: "absolute",
				backgroundColor: "white",
				width: sidebarWidth,
				height: `calc(100% - ${layout.editorToolbarHeight}px)`,
				bottom: 0,
				right: 0,
				overflowY: "hidden",
			}}>
				{this.sidebarFunctionHeader(defData)}
				{(defData && !this.state.previewLocation) && <DefinitionDocumentationHeader defData={defData} />}
				<hr style={dividerSx} />
				<div id={REFERENCES_SECTION_ID}>
					<FlexContainer items="center" style={{ height: 35, padding: whitespace[3] }}>
						<Heading level={7} color="gray" compact={true}>
							References  {refsLoading && <Spinner style={{ marginLeft: whitespace[1] }} />}
						</Heading>
					</FlexContainer>
				</div>
				{refModel === null && <Panel hover={false} hoverLevel="low" style={{
					padding: whitespace[3],
					margin: whitespace[3],
					color: colors.text(),
					textAlign: "center",
				}}>
					<Report width={36} color={colors.blueGrayL1()} /><br />
					We couldn't find any references<br /> for this symbol
				</Panel>}
				{refModel && <RefTree
					model={refModel}
					focus={this.focusResource} />}
				<hr style={dividerSx} />
			</FlexContainer>
			<Preview
				location={this.state.previewLocation || null}
				hidePreview={this.refsFocused} />
		</div>;
	}
};

export const infoStore = new MiniStore<Props | null>();
