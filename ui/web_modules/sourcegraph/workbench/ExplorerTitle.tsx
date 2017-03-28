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
import { URIUtils } from "sourcegraph/core/uri";
import { Features } from "sourcegraph/util/features";
import { onWorkspaceUpdated } from "sourcegraph/workbench/services";
import { Services, getCurrentWorkspace } from "sourcegraph/workbench/services";
import "sourcegraph/workbench/styles/searchViewlet";
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

interface TitleState {
	workspace: IWorkspace;
	openViewlet: string;
}

const buttonSx = {
	flex: "0 0 auto",
	padding: whitespace[1],
	paddingTop: "0.125rem",
	marginRight: 5,
	marginLeft: 5,
};

export class ExplorerTitle extends React.Component<{}, Partial<TitleState>> {

	disposables: IDisposable[] = [];
	commandService: ICommandService = Services.get(ICommandService) as ICommandService;
	viewletService: IViewletService = Services.get(IViewletService) as IViewletService;
	state: TitleState = {
		openViewlet: this.viewletService.getDefaultViewletId(),
		workspace: getCurrentWorkspace(),
	};

	private searchButtonClicked = () => {
		this.updateViewlet(this.state.openViewlet === SEARCH_VIEWLET_ID ? EXPLORER_VIELET_ID : SEARCH_VIEWLET_ID);
	}

	private changesButtonClicked = () => {
		this.updateViewlet(this.state.openViewlet === SCM_VIEWLET_ID ? EXPLORER_VIELET_ID : SCM_VIEWLET_ID);
	}
	private repoNameClicked = () => {
		this.updateViewlet(EXPLORER_VIELET_ID);
	}

	private updateViewlet(viewletId: string): void {
		if (this.state.openViewlet === viewletId) { return; }
		this.setState({ openViewlet: viewletId }, () => {
			this.commandService.executeCommand(viewletId);
		});
	}

	private repoDisplayName(): string {
		const workspace = this.state.workspace;
		if (!workspace) { return ""; }
		const resource = workspace.resource;
		let { repo } = URIUtils.repoParams(resource);
		// for the explorer viewlet, we don't want to show the authority (github.com/)
		return repo.slice(resource.authority.length + 1);
	}

	componentDidMount(): void {
		this.disposables.push(this.viewletService.onDidViewletOpen(v => {
			this.setState({ openViewlet: v.getId() });
		}));
		this.disposables.push(onWorkspaceUpdated(workspace => {
			if (workspace.revState && workspace.revState.zapRef) {
				this.updateViewlet(SCM_VIEWLET_ID);
			} else if (this.state.openViewlet === SCM_VIEWLET_ID) {
				this.updateViewlet(EXPLORER_VIELET_ID);
			}
			this.setState({ workspace });
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
			minHeight: layout.editorToolbarHeight,
			position: "relative",
			paddingLeft: whitespace[2],
			paddingRight: whitespace[2],
			zIndex: 1,
			width: "100%",
		}}>
			<Heading level={6} compact={true} style={{
				lineHeight: 0,
				maxWidth: "85%",
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
				{Features.textSearch.isEnabled() && <Button onClick={this.searchButtonClicked} color={searchMode ? "blue" : "blueGray"}
					{...hover({ backgroundColor: !searchMode ? `${colors.blueGrayD2()} !important` : "" }) }
					style={buttonSx}><Search style={{ top: 0 }} /></Button>}
				{workspace && workspace.revState && workspace.revState.zapRev &&
					<Button onClick={this.changesButtonClicked} color={changesMode ? "blue" : "blueGray"}
						{...hover({ backgroundColor: !changesMode ? `${colors.blueGrayD2()} !important` : "" }) }
						style={buttonSx}><History style={{ top: 0 }} /></Button>
				}
			</div>
		</FlexContainer >;
	}
}
