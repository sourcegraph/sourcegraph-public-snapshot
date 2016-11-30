import { URIUtils } from "sourcegraph/core/uri";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";
import { IDisposable } from "vs/base/common/lifecycle";
import { TPromise } from "vs/base/common/winjs.base";
import { SimpleEditor, SimpleModel } from "vs/editor/browser/standalone/simpleServices";
import { createModel, getModel } from "vs/editor/browser/standalone/standaloneEditor";
import * as editorCommon from "vs/editor/common/editorCommon";
import { IEditorViewState } from "vs/editor/common/editorCommon";
import { IRange } from "vs/editor/common/editorCommon";
import { getCodeEditor } from "vs/editor/common/services/codeEditorService";
import { IEditor, IEditorService, IResourceInput, ITextEditorModel } from "vs/platform/editor/common/editor";

const fetch = singleflightFetch(defaultFetch);

export interface IEditorOpenedEvent {
	model: editorCommon.IModel;
	editor: editorCommon.IEditor;
}

export class EditorService implements IEditorService {
	public _serviceBrand: any;

	private editor?: SimpleEditor;

	// _savedState holds the last view state for each model. It
	// is keyed on model ID.
	private _savedState: Map<string, IEditorViewState> = new Map();

	private _onDidOpenEditor: (e: IEditorOpenedEvent) => void;

	public setEditor(editor: editorCommon.IEditor): void {
		this.editor = new SimpleEditor(editor);
	}

	// An event emitted when the editor jumps to a new model or position therein.
	public onDidOpenEditor(listener: (e: IEditorOpenedEvent) => void): IDisposable {
		if (this._onDidOpenEditor) {
			throw new Error("onDidOpenEditor listener already set");
		}
		this._onDidOpenEditor = listener;
		return { dispose(): void { this._onDidOpenEditor = null; } };
	}

	public openEditor(data: IResourceInput, sideBySide?: boolean): TPromise<IEditor> {
		if (!this.editor) {
			throw new Error(`editor not available`);
		}

		return this.resolveModel(data, false).then(model => {
			if (!this.editor) {
				throw new Error(`editor not available`);
			}

			if (!model) {
				throw new Error(`model not found: ${data.resource.toString()}`);
			}
			const err = model as any;
			if (err.response && err.response.status === 404) {
				throw new Error("404 file not found");
			}

			const codeEditor = getCodeEditor(this.editor);
			const oldModel = codeEditor.getModel();
			if (oldModel && model.id !== oldModel.id) {
				// Save editor state for old model.
				this._savedState.set(oldModel.id, codeEditor.saveViewState());

				codeEditor.setModel(model);

				// Restore editor state.
				const savedState = this._savedState.get(model.id);
				if (savedState) {
					codeEditor.restoreViewState(savedState);
				}
			}

			const selection = data.options && data.options.selection;
			if (selection) {
				if (typeof selection.endLineNumber === "number" && typeof selection.endColumn === "number") {
					codeEditor.setSelection(selection as IRange);
					codeEditor.revealRangeInCenter(selection as IRange);
				} else {
					const pos = {
						lineNumber: selection.startLineNumber,
						column: selection.startColumn,
					};
					codeEditor.setPosition(pos);
					codeEditor.revealPositionInCenter(pos);
				}
			}

			if (this._onDidOpenEditor) {
				this._onDidOpenEditor({ model: model, editor: this.editor._widget });
			}

			return this.editor;
		});
	}

	private resolveModel(data: IResourceInput, refresh?: boolean): TPromise<editorCommon.IModel> {
		const existingModel = getModel(data.resource);
		if (existingModel) {
			return TPromise.as(existingModel);
		}

		const {repo, rev, path} = URIUtils.repoParams(data.resource);
		return TPromise.wrap(
			fetch(`/.api/graphql`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({
					query: `query Content($repo: String, $rev: String, $path: String) {
						root {
							repository(uri: $repo) {
								commit(rev: $rev) {
									commit {
										file(path: $path) {
											content
										}
									}
								}
							}
						}
					}`,
					variables: { repo, rev, path },
				}),
			})
				.then(checkStatus)
				.then(resp => resp.json())
				.then((resp: GQL.IGraphQLResponseRoot) => {
					if (!resp.data) {
						throw new Error("file content not available");
					}
					// Call getModel again in case we lost a race.
					return getModel(data.resource) || createModel(resp.data.root.repository.commit.commit.file.content, getModeByFilename(path), data.resource);
				})
				.catch(err => err)
		);
	}

	public resolveEditorModel(data: IResourceInput, refresh?: boolean): TPromise<ITextEditorModel> {
		if (!this.editor) {
			throw new Error(`editor not available`);
		}
		return this.resolveModel(data, refresh).then((model: editorCommon.IModel) => new SimpleModel(model));
	}
}

// TODO(sqs): Use the built-in ModeService instead of writing our own
// hacky thing to figure out the mode (language) to use for a given
// file. We need to use this for now because our URIs have the file
// path in the URI fragment, which tricks ModeService's detector.
function getModeByFilename(path: string): string {
	if (path.endsWith(".go")) {
		return "go";
	}
	if (path.endsWith(".js") || path.endsWith(".jsx")) {
		return "javascript";
	}
	if (path.endsWith(".ts") || path.endsWith(".tsx")) {
		return "typescript";
	}
	if (path.endsWith(".py")) {
		return "python";
	}
	if (path.endsWith(".html")) {
		return "html";
	}
	if (path.endsWith(".css")) {
		return "css";
	}
	if (path.endsWith(".php")) {
		return "php";
	}
	if (path.endsWith(".java")) {
		return "java";
	}
	if (path.endsWith(".scala")) {
		return "scala";
	}
	if (path.endsWith(".rb")) {
		return "ruby";
	}
	if (path.endsWith(".c") || path.endsWith(".h")) {
		return "c";
	}
	if (path.endsWith(".cpp")) {
		return "cpp";
	}
	if (path.endsWith(".cs")) {
		return "csharp";
	}
	return "plaintext";
}
