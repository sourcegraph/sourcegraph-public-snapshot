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

import { URIUtils } from "sourcegraph/core/uri";

const toStrip = "github.com/";

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
	}

	getTitle(): string {
		const contextService = (this as any).contextService as IWorkspaceContextService;
		const { resource } = contextService.getWorkspace();
		let { repo } = URIUtils.repoParams(resource);
		if (repo.startsWith(toStrip)) {
			repo = repo.substr(toStrip.length);
		}
		return repo;
	}

	public getActions(): IAction[] {
		return [];
	}
}
