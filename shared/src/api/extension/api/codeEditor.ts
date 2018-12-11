import * as clientType from '@sourcegraph/extension-api-types'
import * as sourcegraph from 'sourcegraph'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import { Range } from '../types/range'
import { Selection } from '../types/selection'
import { ExtDocuments } from './documents'

/** @internal */
export class ExtCodeEditor implements sourcegraph.CodeEditor {
    constructor(
        private resource: string,
        public _selections: clientType.Selection[],
        public readonly isActive: boolean,
        private proxy: ClientCodeEditorAPI,
        private documents: ExtDocuments
    ) {}

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

    public setDecorations(_decorationType: null, decorations: sourcegraph.TextDocumentDecoration[]): void {
        this.proxy.$setDecorations(this.resource, decorations.map(fromTextDocumentDecoration))
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
