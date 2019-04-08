import { ProxyResult } from '@sourcegraph/comlink'
import * as clientType from '@sourcegraph/extension-api-types'
import { BehaviorSubject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import { ClientEditorAPI } from '../../client/api/viewComponents/editor'
import { CodeEditorData, EditorDataCommon, EditorId } from '../../client/services/editorService'
import { Range } from '../types/range'
import { Selection } from '../types/selection'
import { createDecorationType } from './decorations'
import { ExtDocuments } from './documents'
import { ExtEditorCommon } from './viewComponents/editor'

const DEFAULT_DECORATION_TYPE = createDecorationType()

/** @internal */
export class ExtCodeEditor extends ExtEditorCommon implements sourcegraph.CodeEditor {
    /** The URI of this editor's document. */
    private resource: string

    constructor(
        data: CodeEditorData & EditorId,
        editorProxy: ProxyResult<ClientEditorAPI>,
        private proxy: ProxyResult<ClientCodeEditorAPI>,
        private documents: ExtDocuments
    ) {
        super(data.editorId, editorProxy)
        this.resource = data.resource
        this.update(data)
    }

    public readonly selectionsChanges = new BehaviorSubject<sourcegraph.Selection[]>([])

    public readonly type = 'CodeEditor'

    public get document(): sourcegraph.TextDocument {
        return this.documents.get(this.resource)
    }

    public get selection(): sourcegraph.Selection | null {
        return this.selectionsChanges.value.length > 0 ? this.selectionsChanges.value[0] : null
    }

    public get selections(): sourcegraph.Selection[] {
        return this.selectionsChanges.value
    }

    public setDecorations(
        decorationType: sourcegraph.TextDocumentDecorationType | null,
        decorations: sourcegraph.TextDocumentDecoration[]
    ): void {
        // Backcompat: extensions developed against an older version of the API
        // may not supply a decorationType
        decorationType = decorationType || DEFAULT_DECORATION_TYPE
        // tslint:disable-next-line: no-floating-promises
        this.proxy.$setDecorations(this.resource, decorationType.key, decorations.map(fromTextDocumentDecoration))
    }

    public update(data: Pick<CodeEditorData, 'selections'> & EditorDataCommon): void {
        this.selectionsChanges.next(data.selections.map(s => Selection.fromPlain(s)))
        super.update(data)
    }

    public toJSON(): any {
        return { type: this.type, document: this.document }
    }
}

function fromTextDocumentDecoration(decoration: sourcegraph.TextDocumentDecoration): clientType.TextDocumentDecoration {
    return {
        ...decoration,
        range: (decoration.range as Range).toJSON(),
    }
}
