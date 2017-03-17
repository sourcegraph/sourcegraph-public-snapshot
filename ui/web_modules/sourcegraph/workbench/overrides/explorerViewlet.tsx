import { hover, insertGlobal } from "glamor";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { Link } from "react-router";
import { IAction } from "vs/base/common/actions";
import { IDisposable } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { IContextKeyService } from "vs/platform/contextkey/common/contextkey";
import { IFileService } from "vs/platform/files/common/files";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { IStorageService } from "vs/platform/storage/common/storage";
import { ITelemetryService } from "vs/platform/telemetry/common/telemetry";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";
import { IEditorGroupService } from "vs/workbench/services/group/common/groupService";
import { ExplorerViewlet as VSExplorerViewlet } from "vscode/src/vs/workbench/parts/files/browser/explorerViewlet";

import { __getRouterForWorkbenchOnly, getRevFromRouter } from "sourcegraph/app/router";
import { FlexContainer, Heading } from "sourcegraph/components";
import { Button } from "sourcegraph/components/Button";
import { List } from "sourcegraph/components/symbols/Primaries";
import { History } from "sourcegraph/components/symbols/Primaries";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { URIUtils } from "sourcegraph/core/uri";
import { urlToRepoRev } from "sourcegraph/repo/routes";
import { workbenchStore } from "sourcegraph/workbench/main";
import { WorkbenchEditorService } from "sourcegraph/workbench/overrides/editorService";
import { onWorkspaceUpdated } from "sourcegraph/workbench/services";
import { RouterContext } from "sourcegraph/workbench/utils";

export class ExplorerViewlet extends VSExplorerViewlet {
	private _editorService: WorkbenchEditorService;
	private _contextService: IWorkspaceContextService;

	constructor(
		@ITelemetryService telemetryService: ITelemetryService,
		@IWorkspaceContextService contextService: IWorkspaceContextService,
		@IStorageService storageService: IStorageService,
		@IEditorGroupService editorGroupService: IEditorGroupService,
		@IWorkbenchEditorService editorService: IWorkbenchEditorService,
		@IConfigurationService configurationService: IConfigurationService,
		@IInstantiationService instantiationService: IInstantiationService,
		@IContextKeyService contextKeyService: IContextKeyService,
		@IFileService fileService: IFileService,
	) {
		super(telemetryService, contextService, storageService, editorGroupService, editorService, configurationService, instantiationService, contextKeyService);

		this._contextService = contextService;
		this._contextService.onWorkspaceUpdated(() => {
			this.updateTitleArea();
		});

		this._editorService = editorService as WorkbenchEditorService;
		this.onTitleAreaUpdate(() => this.updateTitleComponent());
		fileService.onFileChanges(() => this.refresh());
	}

	getTitle(): string {
		const resource = this.getResource();
		let { repo } = URIUtils.repoParams(resource);
		// for the explorer viewlet, we don't want to show the authority (github.com/)
		repo = repo.slice(resource.authority.length + 1);
		return repo;
	}

	getRepoUrl(): string {
		const resource = this.getResource();
		let { repo } = URIUtils.repoParams(resource);
		return repo;
	}

	private getResource(): URI {
		const contextService = (this as any).contextService as IWorkspaceContextService;
		const { resource } = contextService.getWorkspace();
		return resource;
	}

	public getActions(): IAction[] {
		return [];
	}

	public refresh(): void {
		const explorerView = this.getExplorerView();
		explorerView.refresh();
	}

	private updateTitleComponent = (): void => {
		const parent = document.getElementById("workbench.parts.sidebar");
		if (!parent) {
			requestAnimationFrame(this.updateTitleComponent);
			return;
		}
		let titleElement = parent.children[0];
		if (!titleElement || titleElement.className !== "composite title") {
			throw new Error("Wrong element");
		}
		const workspace = this._contextService.getWorkspace();
		ReactDOM.render(<RouterContext>
			<Title repoDisplayName={this.getTitle()} repo={this.getRepoUrl()} revState={workspace.revState} />
		</RouterContext>, titleElement);
	}
}

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

interface TitleProps {
	repoDisplayName: string;
	repo: string;
	revState?: { zapRev?: string, commitID?: string, branch?: string };
}

interface TitleState {
	revState?: { zapRev?: string, commitID?: string, branch?: string };
	diffMode: boolean;
}

class Title extends React.Component<TitleProps, Partial<TitleState>> {
	disposables: IDisposable[];

	constructor(props: TitleProps) {
		super(props);
		this.state = {
			revState: this.props.revState,
			diffMode: workbenchStore.getState().diffMode,
		};
		workbenchStore.subscribe(
			() => this.setState(workbenchStore.getState())
		);
		this.disposables = [];
	}

	componentDidMount(): void {
		this.disposables.push(onWorkspaceUpdated(workspace => {
			this.setState({
				revState: workspace.revState,
			});
		}));
	}

	componentWillUnmount(): void {
		this.disposables.forEach(disposable => disposable.dispose());
	}

	setDiffMode(diffMode: boolean): void {
		workbenchStore.dispatch({ diffMode });
	}

	render(): JSX.Element {
		// TODO(john): make router properties injectable, so this component receives router props and re-renders
		// whenever router props change.
		const router = __getRouterForWorkbenchOnly();
		const rev = getRevFromRouter(router);

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
				<Link to={urlToRepoRev(this.props.repo, rev)}
					{...hover({ color: `${colors.white()} !important` }) }
					style={{
						color: colors.blueGrayL2(),
						maxWidth: "100%",
						overflow: "hidden",
						textOverflow: "ellipsis",
						display: "inline-block",
						marginTop: 5,
					}}>
					<List width={21} style={{ opacity: 0.6, marginRight: whitespace[1] }} />
					{this.props.repoDisplayName}
				</Link>
			</Heading>
			{this.state.revState && this.state.revState.zapRev &&
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
