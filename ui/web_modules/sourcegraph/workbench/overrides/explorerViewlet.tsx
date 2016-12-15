import { IAction } from "vs/base/common/actions";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { ExplorerViewlet as VSExplorerViewlet } from "vs/workbench/parts/files/browser/explorerViewlet";

import { URIUtils } from "sourcegraph/core/uri";

const toStrip = "github.com/";

export class ExplorerViewlet extends VSExplorerViewlet {
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
