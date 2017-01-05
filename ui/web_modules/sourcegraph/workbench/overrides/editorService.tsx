import Event, { Emitter } from "vs/base/common/event";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IEditor } from "vs/platform/editor/common/editor";
import * as vs from "vscode/src/vs/workbench/services/editor/browser/editorService";

export class WorkbenchEditorService extends vs.WorkbenchEditorService {
	private _emitter: Emitter<URI> = new Emitter<URI>();

	public openEditor(data: any, options: any, position?: any): TPromise<IEditor> {
		this._emitter.fire(data.resource);
		return super.openEditor(data, options, position);
	}

	public onDidOpenEditor: Event<URI> = this._emitter.event;
}

export const DelegatingWorkbenchEditorService = vs.DelegatingWorkbenchEditorService;
