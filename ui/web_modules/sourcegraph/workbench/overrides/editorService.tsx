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
import { urlToRepo } from "sourcegraph/repo/routes";
import { prettifyRev } from "sourcegraph/workbench/utils";

export class WorkbenchEditorService extends vs.WorkbenchEditorService {
	private _onDidOpenEditor: Emitter<URI> = new Emitter<URI>();
	private _diffMode: boolean = false;

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

			// openEditor may be called for a file, or for a workspace root (directory);
			// in the latter case, we circumvent the vscode path to open an real document
			// otherwise an empty buffer will be shown instead of the workbench watermark.
			const url = resource.fragment === "" ? urlToRepo(repo) : urlToBlob(repo, rev, path);
			const promise = resource.fragment === "" ? TPromise.wrap(this.getActiveEditor()) : this.openEditorWithoutURLChange(resource, data, options);
			return promise.then(editor => {
				router.push({
					pathname: url,
					state: options,
					hash,
					query: router.location.query,
				});
				return editor;
			});
		}
		throw new Error("cannot open editor, missing resource");
	}

	openEditorWithoutURLChange(mainResource: URI, data: IResourceInput, options?: any): TPromise<IEditor>;
	openEditorWithoutURLChange(mainResource: URI, data: IEditorInput, options?: any): TPromise<IEditor>;
	openEditorWithoutURLChange(mainResource: URI, data: null, options?: any): TPromise<IEditor>;
	openEditorWithoutURLChange(mainResource: URI, data: any, options?: any): TPromise<IEditor> {
		if (!data) {
			data = { resource: mainResource };
		}
		if (data.resource) {
			this._diffMode = false;
		} else if (data.modifiedInput) {
			this._diffMode = true;
		}

		this._onDidOpenEditor.fire(mainResource);

		// calling openEditor with a non-zero position, or options equal to
		// true opens up another editor to the side.
		if (typeof options === "boolean") {
			options = false;
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
			return this.createInput(data).then(input => super.openEditor(input, options, false));
		});
	}

	public createInput(data: IResourceInput): TPromise<IEditorInput>;
	public createInput(data: IEditorInput): TPromise<IEditorInput>;
	public createInput(data: any): TPromise<IEditorInput> {
		if (data.resource && data.resource instanceof URI && data.resource.scheme === "git" && URIUtils.hasAbsoluteCommitID(data.resource)) {
			return TPromise.as((this as any).createFileInput(data.resource)); // access superclass's private method
		}
		return super.createInput(data);
	}

	get diffMode(): boolean {
		return this._diffMode;
	}

	public onDidOpenEditor: Event<URI> = this._onDidOpenEditor.event;
}

export const DelegatingWorkbenchEditorService = vs.DelegatingWorkbenchEditorService;
