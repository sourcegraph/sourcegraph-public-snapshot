import { JointRange } from '../editor'

export interface Completion {
    prefix: string
    content: string
    stopReason?: string
}

export interface TextDocumentAutocompleteContext {
    languageId: string
    markdownLanguage: string

    prefix: JointRange | null
    suffix: JointRange | null

    prevLine: JointRange | null
    prevNonEmptyLine: JointRange | null
    nextNonEmptyLine: JointRange | null
}
