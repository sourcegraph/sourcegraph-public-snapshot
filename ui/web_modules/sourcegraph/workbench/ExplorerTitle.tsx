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
import { workbenchStore } from "sourcegraph/workbench/main";
import { onWorkspaceUpdated } from "sourcegraph/workbench/services";
import { Services, getCurrentWorkspace } from "sourcegraph/workbench/services";
import "sourcegraph/workbench/styles/searchViewlet.css";
import { ICommandService } from "vs/platform/commands/common/commands";

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

const hoverStyle = hover({ color: `${colors.white()} !important` });

interface TitleState {
	workspace: IWorkspace;
	diffMode: boolean;
}

export class ExplorerTitle extends React.Component<{}, Partial<TitleState>> {

	disposables: IDisposable[] = [];
	commandService: ICommandService = Services.get(ICommandService) as ICommandService;
	state: TitleState = {
		...workbenchStore.getState(),
		workspace: getCurrentWorkspace(),
	};

	private showSearchViewlet = () => {
		this.commandService.executeCommand(SEARCH_VIEWLET_ID);
	}

	private showExplorerViewlet = () => {
		this.commandService.executeCommand(EXPLORER_VIELET_ID);
	}

	private repoDisplayName(): string {
		const workspace = this.state.workspace;
		if (!workspace) { return ""; }
		const resource = workspace.resource;
		let { repo } = URIUtils.repoParams(resource);
		let repoName = resource.fsPath.split("/");
		// for the explorer viewlet, we don't want to show the authority (github.com/)
		return repoName[1] || repo.slice(resource.authority.length + 1);
	}

	componentDidMount(): void {
		this.disposables.push(workbenchStore.subscribe(
			(state) => this.setState({ ...state })
		));
		this.disposables.push(onWorkspaceUpdated(workspace => {
			this.setState({ workspace });
		}));
	}

	componentWillUnmount(): void {
		this.disposables.forEach(disposable => disposable.dispose());
	}

	setDiffMode(diffMode: boolean): void {
		workbenchStore.dispatch({ diffMode });
	}

	render(): JSX.Element {
		const { workspace } = this.state;
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
				<a onClick={this.showExplorerViewlet}
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
			{Features.textSearch.isEnabled() && <a {...hoverStyle}
				style={{ marginRight: 5 }}
				onClick={this.showSearchViewlet}>
				<Search color={colors.blueGrayL2()} />
			</a>}
			{workspace && workspace.revState && workspace.revState.zapRev &&
				<Button onClick={() => this.setDiffMode(!this.state.diffMode)} color={this.state.diffMode ? "blue" : "blueGray"}
					{...hover({ backgroundColor: !this.state.diffMode ? `${colors.blueGrayD2()} !important` : "" }) }
					style={{
						flex: "0 0 auto",
						padding: whitespace[1],
						paddingTop: "0.125rem",
					}}><History style={{ top: 0 }} /></Button>
			}
		</FlexContainer >;
	}
}
