import { ProxyResult } from '@sourcegraph/comlink'
import * as clientType from '@sourcegraph/extension-api-types'
import { BehaviorSubject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import { CodeEditorData } from '../../client/services/editorService'
import { Range } from '../types/range'
import { Selection } from '../types/selection'
import { createDecorationType } from './decorations'
import { ExtDocuments } from './documents'

const DEFAULT_DECORATION_TYPE = createDecorationType()

/** @internal */
export class ExtCodeEditor implements sourcegraph.CodeEditor {
    /** The URI of the text document shown in this code editor */
    private resource: string

    constructor(
        data: CodeEditorData,
        private proxy: ProxyResult<ClientCodeEditorAPI>,
        private documents: ExtDocuments
    ) {
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

    public update(data: Pick<CodeEditorData, 'selections'>): void {
        this.selectionsChanges.next(data.selections.map(s => Selection.fromPlain(s)))
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
