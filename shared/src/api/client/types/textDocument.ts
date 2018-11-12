import minimatch from 'minimatch'
import { DocumentFilter, DocumentSelector } from 'sourcegraph'

/**
 * A literal to identify a text document in the client.
 */
export interface TextDocumentIdentifier {
    /**
     * The text document's uri.
     */
    uri: string
}

/**
 * An item to transfer a text document from the client to the server.
 */
export interface TextDocumentItem {
    uri: string
    languageId: string
    text: string
}

export function match(
    selectors: DocumentSelector | IterableIterator<DocumentSelector>,
    document: TextDocumentItem
): boolean {
    for (const selector of isSingleDocumentSelector(selectors) ? [selectors] : selectors) {
        if (match1(selector, document)) {
            return true
        }
    }
    return false
}

function isSingleDocumentSelector(
    value: DocumentSelector | IterableIterator<DocumentSelector>
): value is DocumentSelector {
    return Array.isArray(value) && (value.length === 0 || isDocumentSelectorElement(value[0]))
}

function isDocumentSelectorElement(value: any): value is DocumentSelector[0] {
    return typeof value === 'string' || isDocumentFilter(value)
}

function isDocumentFilter(value: any): value is DocumentFilter {
    const candidate: DocumentFilter = value
    return (
        typeof candidate.language === 'string' ||
        typeof candidate.scheme === 'string' ||
        typeof candidate.pattern === 'string'
    )
}

function match1(selector: DocumentSelector, document: TextDocumentItem): boolean {
    return score(selector, document.uri, document.languageId) !== 0
}

/**
 * Taken from
 * https://github.com/Microsoft/vscode/blob/3d35801127f0a62d58d752bc613506e836c5d120/src/vs/editor/common/modes/languageSelector.ts#L24.
 */
export function score(selector: DocumentSelector, candidateUri: string, candidateLanguage: string): number {
    // array -> take max individual value
    let ret = 0
    for (const filter of selector) {
        const value = score1(filter, candidateUri, candidateLanguage)
        if (value === 10) {
            return value // already at the highest
        }
        if (value > ret) {
            ret = value
        }
    }
    return ret
}

function score1(selector: DocumentSelector[0], candidateUri: string, candidateLanguage: string): number {
    if (typeof selector === 'string') {
        // Shorthand notation: "mylang" -> {language: "mylang"}, "*" -> {language: "*""}.
        if (selector === '*') {
            return 5
        } else if (selector === candidateLanguage) {
            return 10
        } else {
            return 0
        }
    }

    const { language, scheme, pattern } = selector
    let ret = 0
    if (scheme) {
        if (candidateUri.startsWith(scheme + ':')) {
            ret = 10
        } else if (scheme === '*') {
            ret = 5
        } else {
            return 0
        }
    }
    if (language) {
        if (language === candidateLanguage) {
            ret = 10
        } else if (language === '*') {
            ret = Math.max(ret, 5)
        } else {
            return 0
        }
    }
    if (pattern) {
        if (pattern === candidateUri || candidateUri.endsWith(pattern) || minimatch(candidateUri, pattern)) {
            ret = 10
        } else if (minimatch(candidateUri, '**/' + pattern)) {
            ret = 5
        } else {
            return 0
        }
    }
    return ret
}
