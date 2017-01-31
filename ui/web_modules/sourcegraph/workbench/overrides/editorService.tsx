import Event, { Emitter } from "vs/base/common/event";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IEditor } from "vs/platform/editor/common/editor";
import * as vs from "vscode/src/vs/workbench/services/editor/browser/editorService";

import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { urlToBlob } from "sourcegraph/blob/routes";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { updateFileTree } from "sourcegraph/editor/config";
import { fetchContentAndResolveRev } from "sourcegraph/editor/contentLoader";
import { prettifyRev } from "sourcegraph/workbench/utils";

export class WorkbenchEditorService extends vs.WorkbenchEditorService {
	private _emitter: Emitter<URI> = new Emitter<URI>();

	public openEditor(data: any, options?: any): TPromise<IEditor> {
		let { repo, rev, path } = URIUtils.repoParams(data.resource);
		rev = prettifyRev(rev);
		const router = __getRouterForWorkbenchOnly();

		let hash: undefined | string = undefined;
		if (data.options && data.options.selection) {
			const selection = RangeOrPosition.fromMonacoRange(data.options.selection);
			hash = `#L${selection}`;
		}

		const url = urlToBlob(repo, rev, path);
		router.push({
			pathname: url,
			state: options,
			hash,
		});
		return this.openEditorWithoutURLChange(data, options);
	}

	openEditorWithoutURLChange(data: any, options?: any): TPromise<IEditor> {
		this._emitter.fire(data.resource);
		const router = __getRouterForWorkbenchOnly();

		// calling openEditor with a non-zero position, or options equal to
		// true opens up another editor to the side.
		if (typeof options === "boolean") {
			options = false;
		} else if (options === undefined) {
			options = router.location.state;
		}

		// Set the resource revision to the commit hash
		return TPromise.wrap(fetchContentAndResolveRev(data.resource)).then(({ content, commit }) => {
			data.resource = data.resource.with({ query: commit });
			updateFileTree(data.resource);
			return super.openEditor(data, options, 0);
		});
	}

	public onDidOpenEditor: Event<URI> = this._emitter.event;
}

export const DelegatingWorkbenchEditorService = vs.DelegatingWorkbenchEditorService;
