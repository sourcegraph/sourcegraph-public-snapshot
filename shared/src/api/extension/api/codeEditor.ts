import * as clientType from '@sourcegraph/extension-api-types'
import { ProxyResult } from 'comlink'
import { of } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import { Range } from '../types/range'
import { Selection } from '../types/selection'
import { createDecorationType } from './decorations'
import { ExtDocuments } from './documents'

const DEFAULT_DECORATION_TYPE = createDecorationType()

/** @internal */
export class ExtCodeEditor implements sourcegraph.CodeEditor {
    constructor(
        private resource: string,
        public _selections: clientType.Selection[],
        public readonly isActive: boolean,
        private proxy: ProxyResult<ClientCodeEditorAPI>,
        private documents: ExtDocuments
    ) {}

    public readonly selectionsChanges = of(this.selections)

    public readonly type = 'CodeEditor'

    public get document(): sourcegraph.TextDocument {
        return this.documents.get(this.resource)
    }

    public get selection(): sourcegraph.Selection | null {
        return this._selections.length > 0 ? Selection.fromPlain(this._selections[0]) : null
    }

    public get selections(): sourcegraph.Selection[] {
        return this._selections.map(data => Selection.fromPlain(data))
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
