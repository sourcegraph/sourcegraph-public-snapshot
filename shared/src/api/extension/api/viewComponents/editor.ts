import { ProxyResult } from '@sourcegraph/comlink'
import { BehaviorSubject, Subscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ClientEditorAPI } from '../../../client/api/viewComponents/editor'
import { EditorDataCommon, EditorId } from '../../../client/services/editorService'

/**
 * @internal
 * @template D The editor data type.
 */
export abstract class ExtEditorCommon implements Pick<sourcegraph.CodeEditor, 'collapsed' | 'collapsedChanges'> {
    constructor(
        /** The unique ID of this editor. */
        protected readonly editorId: EditorId['editorId'],
        private _proxy: ProxyResult<ClientEditorAPI>
    ) {}

    public abstract readonly type: string

    private _collapsedChanges = new BehaviorSubject<boolean>(false)

    public get collapsed(): boolean {
        return this._collapsedChanges.value
    }

    public set collapsed(value: boolean) {
        this._collapsedChanges.next(Boolean(value))
        // tslint:disable-next-line: no-floating-promises
        this._proxy.$setCollapsed(this.editorId, value)
    }

    public get collapsedChanges(): Subscribable<boolean> {
        return this._collapsedChanges
    }

    /** Subclasses should override this method and then call `super.update(data)`. */
    public update(data: EditorDataCommon): void {
        this._collapsedChanges.next(Boolean(data.collapsed))
    }
}
