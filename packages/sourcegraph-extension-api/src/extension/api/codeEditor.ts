import * as sourcegraph from 'sourcegraph'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import * as plain from '../../protocol/plainTypes'
import { Range } from '../types/range'
import { ExtDocuments } from './documents'

/** @internal */
export class ExtCodeEditor implements sourcegraph.CodeEditor {
    constructor(private resource: string, private proxy: ClientCodeEditorAPI, private documents: ExtDocuments) {}

    public readonly type = 'CodeEditor'

    public get document(): sourcegraph.TextDocument {
        return this.documents.get(this.resource)
    }

    public setDecorations(_decorationType: null, decorations: sourcegraph.TextDocumentDecoration[]): void {
        this.proxy.$setDecorations(this.resource, decorations.map(fromTextDocumentDecoration))
    }

    public toJSON(): any {
        return { type: this.type, document: this.document }
    }
}

function fromTextDocumentDecoration(decoration: sourcegraph.TextDocumentDecoration): plain.TextDocumentDecoration {
    return {
        ...decoration,
        range: (decoration.range as Range).toJSON(),
    }
}
