export interface Completion {
    prefix: string
    content: string
    stopReason?: string
}

export interface CurrentDocumentContext {
    prefix: string
    suffix: string
    prevLine: string
    prevNonEmptyLine: string
    nextNonEmptyLine: string
}

export interface CurrentDocumentContextWithLanguage extends CurrentDocumentContext {
    languageId: string
    markdownLanguage: string
}

export interface TextEditor {
    getOpenDocuments(): LightTextDocument[]
    getCurrentDocument(): LightTextDocument | null
    getDocumentTextTruncated(uri: string): Promise<string | null>
    getDocumentRelativePath(uri: string): Promise<string | null>
    getTabSize(): number
}

export interface LightTextDocument {
    uri: string
    languageId: string
}

export interface History {
    addItem(newItem: LightTextDocument): void
    lastN(n: number, languageId?: string, ignoreUris?: string[]): LightTextDocument[]
}
