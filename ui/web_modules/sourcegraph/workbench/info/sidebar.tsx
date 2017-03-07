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
import { URIUtils } from "sourcegraph/core/uri";
import { Events, FileEventProps } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { DefinitionData } from "sourcegraph/util/RefsBackend";
import { DefinitionDocumentationHeader } from "sourcegraph/workbench/info/documentation";
import { Preview } from "sourcegraph/workbench/info/preview";
import { sidebarWidth } from "sourcegraph/workbench/info/preview";
import { OneReference, ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { Disposables } from "sourcegraph/workbench/utils";
import { MiniStore } from "sourcegraph/workbench/utils";

export const REFERENCES_SECTION_ID = "references-section-header";
const TreeSidebarClassName = "sg-sidebar";

export interface InfoPanelProps {
	repo: GQL.IRepository;
	fileEventProps: FileEventProps;
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
			if (info.prepareData) {
				// Close preview when escape key is clicked - Don't dismiss sidepane.
				if (!info.prepareData.open && !info.id && this.infoPanelRef instanceof InfoPanel && this.infoPanelRef.state.previewLocation) {
					Events.InfoPanelRefPreview_Closed.logEvent(this.infoPanelRef.getEventProps());
					this.infoPanelRef.setState({
						previewLocation: null,
						refModel: this.infoPanelRef.state.refModel,
					});
					return;
				}

				if (info.prepareData.open) {
					// Log sidebar toggling opened
					Events.InfoPanel_Initiated.logEvent(this.props.fileEventProps);
				} else if (this.infoPanel && this.infoPanel.open) {
					// Log sidebar toggling closed
					Events.InfoPanel_Dismissed.logEvent(this.props.fileEventProps);
				}

				this.info = info;
				this.infoPanel = { open: info.prepareData.open, id: info.id };
				this.forceUpdate();
				return;
			}
			if (this.infoPanelRef instanceof InfoPanel) {
				const state: State = {};
				let updateState = false;
				if (info.loadingComplete !== undefined) {
					state.loadingComplete = info.loadingComplete;
					updateState = true;
				}
				if (info.refModel !== undefined) {
					this.info = info;
					state.refModel = info.refModel;
					state.previewLocation = this.infoPanelRef.state.previewLocation;
					updateState = true;
				}
				if (updateState) {
					this.infoPanelRef.setState(state);
					return;
				}
			}

			this.info = info;
			this.forceUpdate();
		}));
	}

	componentWillUnmount(): void {
		this.toDispose.dispose();
	}

	renderOutOfTreeDOMNode(): void {
		let parent = document.getElementById("workbench.editors.files.textFileEditor") as HTMLDivElement;
		if (!parent) {
			parent = document.getElementById("workbench.editors.stringEditor") as HTMLDivElement;
		}
		if (!parent) {
			parent = document.getElementById("workbench.editors.textDiffEditor") as HTMLDivElement;
		}
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
	fileEventProps: FileEventProps;
};

@autobind
class InfoPanel extends React.Component<Props, State> {

	constructor(props: Props) {
		super(props);
		this.state = {
			previewLocation: null,
			refModel: this.props.refModel,
			loadingComplete: this.props.loadingComplete
		};
	}

	getEventProps(): FileEventProps {
		if (this.props.defData && this.props.defData.definition) {
			const uri = this.props.defData.definition.uri;
			const { repo, rev, path } = URIUtils.repoParams(uri);
			return Object.assign({}, this.props.fileEventProps, { defRepo: repo, defRev: rev || "", defPath: path });
		}
		return this.props.fileEventProps;
	}

	private refsFocused(): void {
		Events.InfoPanelRefPreview_Closed.logEvent(Object.assign({}, this.getEventProps()));
		this.setState({
			previewLocation: null,
		});
	}

	private focusResource(loc: OneReference): void {
		if (loc.isCurrentWorkspace) {
			Events.InfoPanelLocalRef_Toggled.logEvent(Object.assign({}, this.getEventProps(), { refRepo: `${loc.uri.authority}${loc.uri.path}`, refPath: loc.uri.fragment, refRev: loc.uri.query || "" }));
		} else {
			Events.InfoPanelExternalRef_Toggled.logEvent(Object.assign({}, this.getEventProps(), { refRepo: `${loc.uri.authority}${loc.uri.path}`, refPath: loc.uri.fragment, refRev: loc.uri.query || "" }));
		}
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
			<div onClick={() => infoStore.dispatch({ defData: null, prepareData: { open: false }, loadingComplete: true, id: this.props.id, fileEventProps: this.getEventProps() })}
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
			<FlexContainer direction="top-bottom" style={{
				position: "absolute",
				backgroundColor: "white",
				width: sidebarWidth,
				height: `calc(100% - ${layout.EDITOR_TITLE_HEIGHT}px)`,
				bottom: 0,
				right: 0,
				overflowY: "hidden",
				zIndex: 10,
			}}>
				{this.sidebarFunctionHeader(defData)}
				{(defData && !this.state.previewLocation) && <DefinitionDocumentationHeader defData={defData} eventProps={this.getEventProps()} />}
				<hr style={dividerSx} />
				<div id={REFERENCES_SECTION_ID}>
					<FlexContainer items="center" style={{ height: 35, padding: whitespace[3] }}>
						<Heading level={7} color="gray" compact={true}>
							References  {refsLoading && <Spinner style={{ marginLeft: whitespace[1] }} />}
						</Heading>
					</FlexContainer>
				</div>
				{loadingComplete && (!refModel || refModel.empty) && <Panel hover={false} hoverLevel="low" style={{
					padding: whitespace[3],
					margin: whitespace[3],
					color: colors.text(),
					textAlign: "center",
				}}>
					<Report width={36} color={colors.blueGrayL1()} /><br />
					We couldn't find any references<br /> for this symbol
				</Panel>}
				{refModel && !refModel.empty && <RefTree
					model={refModel}
					focus={this.focusResource} />}
				<hr style={dividerSx} />
			</FlexContainer>
			<Preview
				location={this.state.previewLocation || null}
				hidePreview={this.refsFocused}
				fileEventProps={this.getEventProps()} />
		</div>;
	}
};

export const infoStore = new MiniStore<Props>();
