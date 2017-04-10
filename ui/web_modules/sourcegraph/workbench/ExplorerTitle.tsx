import { hover, insertGlobal } from "glamor";
import * as React from "react";

import { IDisposable } from "vs/base/common/lifecycle";
import { IWorkspace } from "vs/platform/workspace/common/workspace";
import { VIEWLET_ID as SEARCH_VIEWLET_ID } from "vs/workbench/parts/search/common/constants";

import { FlexContainer, Heading } from "sourcegraph/components";
import { Button } from "sourcegraph/components/Button";
import { List } from "sourcegraph/components/symbols/Primaries";
import { History, Search } from "sourcegraph/components/symbols/Primaries";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { onWorkspaceUpdated } from "sourcegraph/workbench/services";
import { Services, } from "sourcegraph/workbench/services";
import "sourcegraph/workbench/styles/searchViewlet";
import { getCurrentWorkspace, getURIContext } from "sourcegraph/workbench/utils";
import { ICommandService } from "vs/platform/commands/common/commands";
import { IViewletService } from "vs/workbench/services/viewlet/browser/viewlet";

insertGlobal(".composite.title", {
	opacity: "1 !important",
	overflow: "visible !important",
});

insertGlobal(".explorer-viewlet .monaco-tree-row:hover", {
	backgroundColor: `${colors.blueGrayD2(0.5)} !important`,
	color: "white !important",
});

insertGlobal(".explorer-viewlet .monaco-tree-row.focused, .explorer-viewlet .monaco-tree .selected", {
	backgroundColor: `${colors.blue()} !important`,
	color: "white !important",
	fontWeight: "bold",
});

const EXPLORER_VIELET_ID = "workbench.view.explorer";
const SCM_VIEWLET_ID = "workbench.view.scm";

const hoverStyle = hover({ color: `${colors.white()} !important` });

interface State {
	workspace: IWorkspace;
	openViewlet: string;
}

const buttonSx = {
	flex: "0 0 auto",
	padding: whitespace[1],
	paddingTop: "0.125rem",
	marginRight: 5,
};

export class ExplorerTitle extends React.Component<{}, State> {

	disposables: IDisposable[] = [];
	commandService: ICommandService = Services.get(ICommandService) as ICommandService;
	viewletService: IViewletService = Services.get(IViewletService) as IViewletService;

	constructor() {
		super();
		const workspace = getCurrentWorkspace();
		this.state = {
			workspace,
			openViewlet: workspace.revState && workspace.revState.zapRef ? SCM_VIEWLET_ID : EXPLORER_VIELET_ID,
		};
	}

	private searchButtonClicked = () => {
		this.updateViewlet(this.state.openViewlet === SEARCH_VIEWLET_ID ? EXPLORER_VIELET_ID : SEARCH_VIEWLET_ID);
	}

	private changesButtonClicked = () => {
		this.updateViewlet(this.state.openViewlet === SCM_VIEWLET_ID ? EXPLORER_VIELET_ID : SCM_VIEWLET_ID);
	}
	private repoNameClicked = () => {
		this.updateViewlet(EXPLORER_VIELET_ID);
	}

	private updateViewlet(viewletId: string, force?: boolean): void {
		switch (viewletId) {
			case EXPLORER_VIELET_ID:
				Events.FileTreeViewlet_Toggled.logEvent();
				break;
			case SEARCH_VIEWLET_ID:
				Events.InRepoSearchViewlet_Toggled.logEvent();
				break;
			case SCM_VIEWLET_ID:
				Events.ChangesViewlet_Toggled.logEvent();
				break;
		}

		if (!force && this.state.openViewlet === viewletId) { return; }
		this.setState({ openViewlet: viewletId } as State, () => {
			this.commandService.executeCommand(viewletId);
		});
	}

	private repoDisplayName(): string {
		const workspace = this.state.workspace;
		if (!workspace) { return ""; }
		const resource = workspace.resource;
		let { repo } = getURIContext(resource);
		let repoParts = repo.split("/");
		return repoParts[repoParts.length - 1];
	}

	componentDidMount(): void {
		// If the initial state is not explorer viewlet, update here.
		this.updateViewlet(this.state.openViewlet, true);
		this.disposables.push(onWorkspaceUpdated(workspace => {
			if (workspace.revState && workspace.revState.zapRef) {
				this.updateViewlet(SCM_VIEWLET_ID);
			} else if (this.state.openViewlet === SCM_VIEWLET_ID) {
				// If the current viewlet is SCM, revert back to EXPLORER.
				this.updateViewlet(EXPLORER_VIELET_ID);
			}
			this.setState({ workspace } as State);
		}));
	}

	componentWillUnmount(): void {
		this.disposables.forEach(disposable => disposable.dispose());
	}

	render(): JSX.Element {
		const { workspace, openViewlet } = this.state;
		const searchMode = openViewlet === SEARCH_VIEWLET_ID;
		const changesMode = openViewlet === SCM_VIEWLET_ID;
		return <FlexContainer items="center" justify="between" style={{
			backgroundColor: colors.blueGrayD1(),
			boxShadow: `0 0 8px 1px ${colors.black(0.25)}`,
			minHeight: layout.EDITOR_TITLE_HEIGHT,
			position: "relative",
			paddingLeft: whitespace[2],
			paddingRight: whitespace[2],
			zIndex: 1,
			width: "100%",
		}}>
			<Heading level={6} compact={true} style={{
				lineHeight: 0,
				maxWidth: "74%",
				whiteSpace: "nowrap",
			}}>
				<a onClick={this.repoNameClicked}
					{...hoverStyle}
					style={{
						color: colors.blueGrayL2(),
						maxWidth: "100%",
						overflow: "hidden",
						textOverflow: "ellipsis",
						display: "inline-block",
						marginTop: 5,
					}}>
					<List width={21} style={{ opacity: 0.6, marginRight: whitespace[1] }} />
					{this.repoDisplayName()}
				</a>
			</Heading>
			<div>
				<Button
					onClick={this.searchButtonClicked}
					color={"blue"}
					{...hover({ backgroundColor: !searchMode ? `${colors.blueGrayD2()} !important` : "transparent" }) }
					style={buttonSx}
					backgroundColor={searchMode ? undefined : "transparent"}
					animation={false}>
					<Search style={{ top: 0 }} />
				</Button>
				{workspace && workspace.revState && workspace.revState.zapRev &&
					<Button
						onClick={this.changesButtonClicked}
						color={"blue"}
						{...hover({ backgroundColor: !changesMode ? `${colors.blueGrayD2()} !important` : "" }) }
						style={buttonSx}
						backgroundColor={changesMode ? "auto" : "transparent"}
						animation={false}>
						<History style={{ top: 0 }} />
					</Button>
				}
			</div>
		</FlexContainer >;
	}
}
