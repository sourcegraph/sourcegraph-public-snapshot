import { hover, insertGlobal } from "glamor";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { Link } from "react-router";
import { IAction } from "vs/base/common/actions";
import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { IContextKeyService } from "vs/platform/contextkey/common/contextkey";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { IStorageService } from "vs/platform/storage/common/storage";
import { ITelemetryService } from "vs/platform/telemetry/common/telemetry";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";
import { IEditorGroupService } from "vs/workbench/services/group/common/groupService";
import { ExplorerViewlet as VSExplorerViewlet } from "vscode/src/vs/workbench/parts/files/browser/explorerViewlet";

import { FlexContainer, Heading } from "sourcegraph/components";
import { List } from "sourcegraph/components/symbols/Primaries";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { URIUtils } from "sourcegraph/core/uri";
import { urlToRepo } from "sourcegraph/repo/routes";
import { RouterContext } from "sourcegraph/workbench/utils";

export class ExplorerViewlet extends VSExplorerViewlet {

	constructor(
		@ITelemetryService telemetryService: ITelemetryService,
		@IWorkspaceContextService contextService: IWorkspaceContextService,
		@IStorageService storageService: IStorageService,
		@IEditorGroupService editorGroupService: IEditorGroupService,
		@IWorkbenchEditorService editorService: IWorkbenchEditorService,
		@IConfigurationService configurationService: IConfigurationService,
		@IInstantiationService instantiationService: IInstantiationService,
		@IContextKeyService contextKeyService: IContextKeyService
	) {
		super(telemetryService, contextService, storageService, editorGroupService, editorService, configurationService, instantiationService, contextKeyService);

		contextService.onWorkspaceUpdated(() => {
			this.updateTitleArea();
		});

		this.onTitleAreaUpdate(() => this.updateTitleComponent());
	}

	getTitle(): string {
		const contextService = (this as any).contextService as IWorkspaceContextService;
		const { resource } = contextService.getWorkspace();
		let { repo } = URIUtils.repoParams(resource);
		return repo;
	}

	public getActions(): IAction[] {
		return [];
	}

	private updateTitleComponent(): void {
		const parent = document.getElementById("workbench.parts.sidebar");
		if (!parent) {
			throw new Error("Expected SideBar element to exist.");
		}
		let titleElement = parent.children[0];
		if (!titleElement || titleElement.className !== "composite title") {
			throw new Error("Wrong element");
		}
		ReactDOM.render(<RouterContext>
			<Title repo={this.getTitle()} />
		</RouterContext>, titleElement);
	}
}

function Title({repo}: { repo: string }): JSX.Element {

	insertGlobal(".composite.title", {
		opacity: "1 !important",
		overflow: "visible !important",
	});

	return <FlexContainer items="center" style={{
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
			marginTop: 2,
			maxWidth: "100%",
			whiteSpace: "nowrap",
		}}>
			<Link to={urlToRepo(repo)}
				{...hover({ color: `${colors.white()} !important` }) }
				style={{
					color: colors.blueGrayL2(),
					maxWidth: "100%",
					overflow: "hidden",
					textOverflow: "ellipsis",
					display: "inline-block",
				}}>
				<List width={21} style={{ opacity: 0.6, marginRight: whitespace[1] }} />
				{repo.replace(/^github.com\//, "")}
			</Link>
		</Heading>
	</FlexContainer>;
}
