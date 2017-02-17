import Event, { Emitter } from "vs/base/common/event";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IEditor, IEditorInput, IResourceInput } from "vs/platform/editor/common/editor";
import * as vs from "vscode/src/vs/workbench/services/editor/browser/editorService";

import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { urlToBlob } from "sourcegraph/blob/routes";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { updateFileTree } from "sourcegraph/editor/config";
import { fetchContentAndResolveRev } from "sourcegraph/editor/contentLoader";
import { Features } from "sourcegraph/util/features";
import { prettifyRev } from "sourcegraph/workbench/utils";

export class WorkbenchEditorService extends vs.WorkbenchEditorService {
	private _emitter: Emitter<URI> = new Emitter<URI>();

	public openEditor(data: IResourceInput, options?: any): TPromise<IEditor>;
	public openEditor(data: IEditorInput, options?: any): TPromise<IEditor>;
	public openEditor(data: any, options?: any): TPromise<IEditor> {
		let resource: URI;
		if (data.resource) {
			resource = data.resource;
		} else if (data.modifiedInput) {
			resource = data.modifiedInput.resource;
		} else {
			throw new Error(`unknown data: ${data}`);
		}
		if (resource) {
			let { repo, rev, path } = URIUtils.repoParams(resource);
			rev = prettifyRev(rev);
			const router = __getRouterForWorkbenchOnly();

			let hash: undefined | string;
			if (data.options && data.options.selection) {
				const selection = RangeOrPosition.fromMonacoRange(data.options.selection);
				hash = `#L${selection}`;
			}

			const url = urlToBlob(repo, rev, path);
			router.push({
				pathname: url,
				state: options,
				hash,
				query: router.location.query,
			});
		}
		return this.openEditorWithoutURLChange(resource, data, options);
	}

	openEditorWithoutURLChange(mainResource: URI, data: IResourceInput, options?: any): TPromise<IEditor>;
	openEditorWithoutURLChange(mainResource: URI, data: IEditorInput, options?: any): TPromise<IEditor>;
	openEditorWithoutURLChange(mainResource: URI, data: null, options?: any): TPromise<IEditor>;
	openEditorWithoutURLChange(mainResource: URI, data: any, options?: any): TPromise<IEditor> {
		if (!data) {
			data = { resource: mainResource };
		}

		this._emitter.fire(mainResource);
		const router = __getRouterForWorkbenchOnly();

		// calling openEditor with a non-zero position, or options equal to
		// true opens up another editor to the side.
		if (typeof options === "boolean") {
			options = false;
		} else if (options === undefined) {
			options = router.location.state;
		}

		// Set the resource revision to the commit hash
		return TPromise.wrap(fetchContentAndResolveRev(mainResource)).then(({ content, commit }) => {
			const absMainResource = mainResource.with({ query: commit });
			if (data.resource) {
				data.resource = absMainResource;
			} else if (data.modifiedInput) {
				data.modifiedInput.resource = absMainResource;
			} else {
				throw new Error(`unknown data: ${data}`);
			}
			updateFileTree(mainResource);
			return this.createInput(data).then(input => super.openEditor(input, options, 0));
		});
	}

	public createInput(data: IResourceInput): TPromise<IEditorInput>;
	public createInput(data: IEditorInput): TPromise<IEditorInput>;
	public createInput(data: any): TPromise<IEditorInput> {
		if (Features.zap2Way.isEnabled() && data.resource && data.resource instanceof URI && data.resource.scheme === "git" && URIUtils.hasAbsoluteCommitID(data.resource)) {
			return TPromise.as((this as any).createFileInput(data.resource)); // access superclass's private method
		}
		return super.createInput(data);
	}

	public onDidOpenEditor: Event<URI> = this._emitter.event;
}

export const DelegatingWorkbenchEditorService = vs.DelegatingWorkbenchEditorService;
