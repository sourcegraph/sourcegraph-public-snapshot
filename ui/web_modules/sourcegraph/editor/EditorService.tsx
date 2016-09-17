// tslint:disable typedef ordered-imports
import {TreeEntry} from "sourcegraph/api";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import {treeEntryFromUri} from "sourcegraph/editor/FileModel";
import { makeRepoRev } from "sourcegraph/repo";

const fetch = singleflightFetch(defaultFetch);

export interface IEditorOpenedEvent {
	model: monaco.editor.IModel;
	editor: monaco.editor.IEditor;
}

export class EditorService implements IEditorService {
	public _serviceBrand: any;

	private editor?: monaco.editor.ICommonCodeEditor;

	// _savedState holds the last view state for each model. It
	// is keyed on model ID.
	private _savedState: Map<string, monaco.editor.IEditorViewState> = new Map();

	private _onDidOpenEditor: (e: IEditorOpenedEvent) => void;

	public setEditor(editor: monaco.editor.ICommonCodeEditor): void {
		this.editor = editor;
	}

	// An event emitted when the editor jumps to a new model or position therein.
	public onDidOpenEditor(listener: (e: IEditorOpenedEvent) => void): monaco.IDisposable {
		if (this._onDidOpenEditor) {
			throw new Error("onDidOpenEditor listener already set");
		}
		this._onDidOpenEditor = listener;
		return {dispose(): void { this._onDidOpenEditor = null; }};
	}

	public openEditor(data: IResourceInput, sideBySide?: boolean): monaco.Promise<monaco.editor.IEditor> {
		if (!this.editor) {
			throw new Error(`editor not available`);
		}

		return this.resolveEditorModel(data, false).then(model => {
			if (!this.editor) {
				throw new Error(`editor not available`);
			}

			if (!model) {
				throw new Error(`model not found: ${data.resource.toString()}`);
			}

			const oldModel = this.editor.getModel();
			if (model.textEditorModel.id !== oldModel.id) {
				// Save editor state for old model.
				this._savedState.set(oldModel.id, this.editor.saveViewState());

				this.editor.setModel(model.textEditorModel);

				// Restore editor state.
				const savedState = this._savedState.get(model.textEditorModel.id);
				if (savedState) {
					this.editor.restoreViewState(savedState);
				}
			}

			const selection = data.options && data.options.selection;
			if (selection) {
				if (typeof selection.endLineNumber === "number" && typeof selection.endColumn === "number") {
					this.editor.setSelection(selection as monaco.IRange);
					this.editor.revealRangeInCenter(selection as monaco.IRange);
				} else {
					const pos = {
						lineNumber: selection.startLineNumber,
						column: selection.startColumn,
					};
					this.editor.setPosition(pos);
					this.editor.revealPositionInCenter(pos);
				}
			}
			this.editor.focus();

			if (this._onDidOpenEditor) {
				this._onDidOpenEditor({model: model.textEditorModel, editor: this.editor});
			}

			(this.editor as any).getControl = () => this.editor; // HACK
			return this.editor;
		}, err => console.error("Error", err));
	}

	public resolveEditorModel(data: IResourceInput, refresh?: boolean): monaco.Promise<ITextEditorModel> {
		if (!this.editor) {
			throw new Error(`editor not available`);
		}

		// HACK
		const hackModel = (m: monaco.editor.IModel): ITextEditorModel => {
			const x: any = {textEditorModel: m};
			x.textEditorModel.textEditorModel = x.textEditorModel;
			return x;
		};

		const existingModel = monaco.editor.getModel(data.resource);
		if (existingModel) {
			return monaco.Promise.as(hackModel(existingModel));
		}

		const {repo, rev, path} = treeEntryFromUri(data.resource);
		return new monaco.Promise((c, e) => {
			fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/tree/${path}?ContentsAsString=true&NoSrclibAnns=true`)
				.then(checkStatus)
				.then(resp => resp.json())
				.then((treeEntry: TreeEntry) => {
					// Call getModel again in case we lost a race.
					const newModel = monaco.editor.getModel(data.resource) || monaco.editor.createModel(treeEntry.ContentsString || "", undefined, data.resource);
					c(hackModel(newModel));
				})
				.catch(err => e(err));
		});
	}
}

// The below interfaces were copied from vscode.

export interface IEditorService {
	_serviceBrand: any;
	openEditor(input: IResourceInput, sideBySide?: boolean): monaco.Promise<monaco.editor.IEditor>;
	resolveEditorModel(input: IResourceInput, refresh?: boolean): monaco.Promise<ITextEditorModel>;
}

export interface IResourceInput {
	resource: monaco.Uri;
	mime?: string;
	encoding?: string;
	options?: ITextEditorOptions;
}

export interface IEditorOptions {
	preserveFocus?: boolean;
	forceOpen?: boolean;
	revealIfVisible?: boolean;
	pinned?: boolean;
	index?: number;
	inactive?: boolean;
}

export interface ITextEditorOptions extends IEditorOptions {
	selection?: {
		startLineNumber: number;
		startColumn: number;
		endLineNumber?: number;
		endColumn?: number;
	};
}

export interface IEditorModel {
	onDispose: monaco.IEvent<void>;
	load(): monaco.Promise<IEditorModel>;
	dispose(): void;
}

export interface ITextEditorModel extends IEditorModel {
	textEditorModel: monaco.editor.IModel;
}
