import { ProxyResult } from '@sourcegraph/comlink'
import { BehaviorSubject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ClientEditorAPI } from '../../../client/api/viewComponents/editor'
import { DiffEditorData, EditorDataCommon, EditorId } from '../../../client/services/editorService'
import { ExtDocuments } from '../documents'
import { ExtEditorCommon } from './editor'

/** @internal */
export class ExtDiffEditorViewComponent extends ExtEditorCommon implements sourcegraph.DiffEditor {
    /** The URI of this editor's original document (on the left-hand side). */
    private originalResource: string

    /** The URI of this editor's modified document (on the right-hand side). */
    private modifiedResource: string

    constructor(
        data: DiffEditorData & EditorId,
        editorProxy: ProxyResult<ClientEditorAPI>,
        private documents: ExtDocuments
    ) {
        super(data.editorId, editorProxy)
        this.originalResource = data.originalResource
        this.modifiedResource = data.modifiedResource
        this.update(data)
    }

    public readonly type = 'DiffEditor'

    public get originalDocument(): sourcegraph.TextDocument {
        return this.documents.get(this.originalResource)
    }

    public get modifiedDocument(): sourcegraph.TextDocument {
        return this.documents.get(this.modifiedResource)
    }

    public get rawDiff(): string | undefined {
        return this.rawDiffChanges.value
    }

    public readonly rawDiffChanges = new BehaviorSubject<string | undefined>(undefined)

    public update(data: Pick<DiffEditorData, 'rawDiff'> & EditorDataCommon): void {
        this.rawDiffChanges.next(data.rawDiff)
        super.update(data)
    }

    public toJSON(): any {
        return {
            type: this.type,
            originalDocument: this.originalDocument,
            modifiedDocument: this.modifiedDocument,
        }
    }
}
