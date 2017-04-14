import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IEditor, IEditorInput, IEditorOptions, IResourceDiffInput, IResourceInput, IResourceSideBySideInput } from "vs/platform/editor/common/editor";
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
	openEditorWithoutURLChange(resource: URI, input: any, options?: any): TPromise<IEditor> {
		if (!input) {
			input = { resource: resource };
		}

		// Set the resource revision to the commit hash
		return TPromise.wrap(resolveRev(resource)).then(({ commit }) => {
			updateFileTree(resource);
			return super.createInput(input).then(i => {
				return super.openEditor(i, options).then((editor) => {
					// Resolve the editor contents again since the contents may have changed.
					// When contents change without triggering a refresh then already open editors
					// will have outdated content.
					i.resolve(true);
					return editor;
				});
			});
		});
	}

	openEditor(input: IEditorInput, options?: IEditorOptions): TPromise<IEditor>;
	openEditor(input: IResourceInput | IResourceDiffInput | IResourceSideBySideInput): TPromise<IEditor>;
	openEditor(input: any, options?: any): TPromise<IEditor> {
		let resource: URI;
		let inputToOpen: any; // we translate the provided input into another, depending on context
		if (input.resource) {
			// IResourceInput, e.g. from clicking on a file from the Explorer vielet.
			const workspace = getCurrentWorkspace();
			if (workspace.revState && workspace.revState.zapRev) {
				// Use a diff input, even though a normal input was requested.
				const resolverService = Services.get(ITextModelResolverService);
				const leftInput = new ResourceEditorInput("", "", getGitBaseResource(input.resource), resolverService);
				const rightInput = new ResourceEditorInput("", "", input.resource, resolverService);
				const diffInput = new DiffEditorInput("", "", leftInput, rightInput);
				resource = input.resource;
				inputToOpen = diffInput;
			} else {
				resource = input.resource;
				inputToOpen = input;
			}
		} else if (input.rightResource) {
			// IResourceDiffInput, e.g. from clicking on a file from the SCM viewlet.
			// Can get here via clicking on a file from "Changes" view or by clicking on file from explorer view
			// while viewing a zap ref.
			const resolverService = Services.get(ITextModelResolverService);
			const leftInput = new ResourceEditorInput("", "", input.leftResource, resolverService);
			const rightInput = new ResourceEditorInput("", "", input.rightResource, resolverService);
			const diffInput = new DiffEditorInput("", "", leftInput, rightInput);
			resource = input.rightResource;
			inputToOpen = diffInput;
		} else {
			throw new Error(`unknown input: ${input}`);
		}

		if (resource) {
			let { repo, rev, path } = getURIContext(resource);
			rev = prettifyRev(rev);
			const router = __getRouterForWorkbenchOnly();

			// openEditor may be called for a file or for a workspace root (directory).
			// In the latter case, we circumvent the vscode path to open a document.
			// Otherwise an empty buffer will be shown instead of the workbench watermark.
			const url = getURIContext(resource).path === "" ? urlToRepo(repo) : urlToBlob(repo, rev, path);
			const promise = getURIContext(resource).path === "" ?
				TPromise.wrap(this.getActiveEditor()) :
				this.openEditorWithoutURLChange(resource, inputToOpen, { ...options, ...input.options });
			return promise.then(editor => {
				router.push({
					pathname: url,
					hash: input.options && input.options.selection ? `#L${RangeOrPosition.fromMonacoRange(input.options.selection)}` : undefined,
					query: router.location.query,
				});
				return editor;
			});
		}
		throw new Error("cannot open editor, unable to determine resource");
	}
}

// This export required by InstantiationService
export const DelegatingWorkbenchEditorService = vs.DelegatingWorkbenchEditorService;

let EditorInstance: ICodeEditor | null = null;

export function getEditorInstance(): ICodeEditor | null {
	return EditorInstance;
}

export function updateEditorInstance(editor: ICodeEditor): void {
	EditorInstance = editor;
}
