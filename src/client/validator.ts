import { TextDocumentDecoration } from '../protocol/decoration'

export interface Validator {
    asTextDocumentDecoration(value: TextDocumentDecoration): TextDocumentDecoration

    asTextDocumentDecorations(values: TextDocumentDecoration[]): TextDocumentDecoration[]
    asTextDocumentDecorations(values: undefined | null): null
    asTextDocumentDecorations(values: TextDocumentDecoration[] | undefined | null): TextDocumentDecoration[] | null
}

export function createValidator(): Validator {
    function asTextDocumentDecoration(value: TextDocumentDecoration): TextDocumentDecoration {
        return value
    }

    function asTextDocumentDecorations(values: TextDocumentDecoration[]): TextDocumentDecoration[]
    function asTextDocumentDecorations(values: undefined | null): null
    function asTextDocumentDecorations(
        values: TextDocumentDecoration[] | undefined | null
    ): TextDocumentDecoration[] | null {
        if (!values) {
            return null
        }
        return values.map(asTextDocumentDecoration)
    }

    return {
        asTextDocumentDecoration,
        asTextDocumentDecorations,
    }
}
