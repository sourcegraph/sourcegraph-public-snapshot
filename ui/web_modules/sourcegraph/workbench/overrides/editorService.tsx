import Event, { Emitter } from "vs/base/common/event";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IEditor, IEditorInput, IResourceInput } from "vs/platform/editor/common/editor";
import { DiffEditorInput } from "vs/workbench/common/editor/diffEditorInput";
import { ResourceEditorInput } from "vs/workbench/common/editor/resourceEditorInput";
import * as vs from "vscode/src/vs/workbench/services/editor/browser/editorService";

import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { urlToBlob } from "sourcegraph/blob/routes";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { updateFileTree } from "sourcegraph/editor/config";
import { resolveRev } from "sourcegraph/editor/contentLoader";
import { urlToRepo } from "sourcegraph/repo/routes";
import { Services } from "sourcegraph/workbench/services";
import { getCurrentWorkspace, getGitBaseResource, getURIContext, prettifyRev } from "sourcegraph/workbench/utils";

export class WorkbenchEditorService extends vs.WorkbenchEditorService {
	private _onDidOpenEditor: Emitter<URI> = new Emitter<URI>();

	// TODO(john): this type signature diverges from vscode's.
	// Really this whole file is a sin...
	public openEditor(data: IResourceInput, options?: any): TPromise<IEditor>;
	public openEditor(data: IEditorInput, options?: any): TPromise<IEditor>;
	public openEditor(data: any, options?: any): TPromise<IEditor> {
		let resource: URI;
		let input: any;
		if (data.resource) {
			const workspace = getCurrentWorkspace();
			if (workspace.revState && workspace.revState.zapRev) {
				const resolverService = Services.get(ITextModelResolverService);
				const leftInput = new ResourceEditorInput("", "", getGitBaseResource(data.resource), resolverService);
				const rightInput = new ResourceEditorInput("", "", data.resource, resolverService);
				const diffInput = new DiffEditorInput("", "", leftInput, rightInput);
				resource = data.resource;
				input = diffInput;
			} else {
				resource = data.resource;
				input = data;
			}
		} else if (data.modifiedInput) {
			resource = data.modifiedInput.resource;
			input = data;
		} else if (data.rightResource) {
			// Can get here via clicking on a file from "Changes" view or by clicking on file from explorer view
			// while viewing a zap ref.
			const resolverService = Services.get(ITextModelResolverService);
			const leftInput = new ResourceEditorInput("", "", data.leftResource, resolverService);
			const rightInput = new ResourceEditorInput("", "", data.rightResource, resolverService);
			const diffInput = new DiffEditorInput("", "", leftInput, rightInput);
			resource = data.rightResource;
			input = diffInput;
		} else {
			throw new Error(`unknown data: ${data}`);
		}
		if (resource) {
			let { repo, rev, path } = getURIContext(resource);
			rev = prettifyRev(rev);
			const router = __getRouterForWorkbenchOnly();

			// openEditor may be called for a file, or for a workspace root (directory);
			// in the latter case, we circumvent the vscode path to open an real document
			// otherwise an empty buffer will be shown instead of the workbench watermark.
			const url = getURIContext(resource).path === "" ? urlToRepo(repo) : urlToBlob(repo, rev, path);
			const promise = getURIContext(resource).path === "" ? TPromise.wrap(this.getActiveEditor()) : this.openEditorWithoutURLChange(resource, input, options);
			return promise.then(editor => {
				router.push({
					pathname: url,
					state: options,
					hash: data.options && data.options.selection ? `#L${RangeOrPosition.fromMonacoRange(data.options.selection)}` : undefined,
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

		this._onDidOpenEditor.fire(mainResource); // TODO(john): why are we firing this here?

		// calling openEditor with a non-zero position, or options equal to
		// true opens up another editor to the side.
		if (typeof options === "boolean") {
			options = false;
		}

		// Set the resource revision to the commit hash
		return TPromise.wrap(resolveRev(mainResource)).then(({ commit }) => {
			updateFileTree(mainResource);
			return this.createInput(data).then(input => super.openEditor(input, options, false));
		});
	}

	public createInput(data: IResourceInput): TPromise<IEditorInput>;
	public createInput(data: IEditorInput): TPromise<IEditorInput>;
	public createInput(data: any): TPromise<IEditorInput> {
		const resource = data.resource;
		if (resource && resource instanceof URI && resource.scheme === "git") {
			return TPromise.as((this as any).createFileInput(resource)); // access superclass's private method
		}
		return super.createInput(data);
	}

	public onDidOpenEditor: Event<URI> = this._onDidOpenEditor.event;
}

export const DelegatingWorkbenchEditorService = vs.DelegatingWorkbenchEditorService;

let EditorInstance: ICodeEditor | null = null;

export function getEditorInstance(): ICodeEditor | null {
	return EditorInstance;
}

export function updateEditorInstance(editor: ICodeEditor): void {
	EditorInstance = editor;
}
