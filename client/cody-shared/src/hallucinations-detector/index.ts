import { marked } from 'marked'

import { parseMarkdown } from '../chat/markdown'

export interface HighlightedToken {
    type: 'file' | 'symbol'
    // Including leading/trailing whitespaces or quotes.
    outerValue: string
    innerValue: string
    isHallucinated: boolean
}

interface HighlightTokensResult {
    text: string
    tokens: HighlightedToken[]
}

export async function highlightTokens(
    text: string,
    fileExists: (filePath: string) => Promise<boolean>
): Promise<HighlightTokensResult> {
    const markdownTokens = parseMarkdown(text)
    const tokens = await detectTokens(markdownTokens, fileExists)

    const highlightedText = markdownTokens
        .map(token => {
            switch (token.type) {
                case 'code':
                case 'codespan':
                    return token.raw
                default:
                    return highlightLine(token.raw, tokens)
            }
        })
        .join('')

    return { text: highlightedText, tokens }
}

async function detectTokens(
    tokens: marked.Token[],
    fileExists: (filePath: string) => Promise<boolean>
): Promise<HighlightedToken[]> {
    const highlightedTokens: HighlightedToken[] = []
    for (const token of tokens) {
        switch (token.type) {
            case 'code':
            case 'codespan':
                continue
            default: {
                const lineTokens = await detectFilePaths(token.raw, fileExists)
                highlightedTokens.push(...lineTokens)
            }
        }
    }
    return deduplicateTokens(highlightedTokens)
}

function highlightLine(line: string, tokens: HighlightedToken[]): string {
    let highlightedLine = line
    for (const token of tokens) {
        highlightedLine = highlightedLine.replaceAll(token.innerValue, getHighlightedTokenHTML(token))
    }
    return highlightedLine
}

function getHighlightedTokenHTML(token: HighlightedToken): string {
    const isHallucinatedClassName = token.isHallucinated ? 'hallucinated' : 'not-hallucinated'
    return `<span class="token-${token.type} token-${isHallucinatedClassName}">${token.innerValue}</span>`
}

function deduplicateTokens(tokens: HighlightedToken[]): HighlightedToken[] {
    const deduplicatedTokens: HighlightedToken[] = []
    const values = new Set<string>()
    for (const token of tokens) {
        if (!values.has(token.outerValue)) {
            deduplicatedTokens.push(token)
            values.add(token.outerValue)
        }
    }
    return deduplicatedTokens
}

function detectFilePaths(
    line: string,
    fileExists: (filePath: string) => Promise<boolean>
): Promise<HighlightedToken[]> {
    const filePaths = Array.from(line.matchAll(filePathRegexp))
        .map(match => ({ fullMatch: match[0], pathMatch: match[1] }))
        .filter(match => isFilePathLike(match.fullMatch, match.pathMatch))
        .map(async (match): Promise<HighlightedToken> => {
            const exists = await fileExists(match.pathMatch)
            return { type: 'file', outerValue: match.fullMatch, innerValue: match.pathMatch, isHallucinated: !exists }
        })
    return Promise.all(filePaths)
}

const filePathCharacters = '[\\w\\/\\._-]'

const filePathRegexpParts = [
    // File path can start with a `, ", ', or a whitespace
    '[`"\'\\s]?',
    // Capture a file path-like sequence.
    `(\\/?${filePathCharacters}+\\/${filePathCharacters}+)`,
    //  File path can end with a `, ", ', ., or a whitespace.
    '[`"\'\\s\\.]?',
]

const filePathRegexp = new RegExp(filePathRegexpParts.join(''), 'g')

function isFilePathLike(fullMatch: string, pathMatch: string): boolean {
    const parts = pathMatch.split(/[/\\]/)

    if (fullMatch.startsWith(' ') && parts.length <= 2) {
        // Probably a / used as an "or" in a sentence. For example, "This is a cool/awesome function."
        return false
    }

    if (parts[0].includes('.com') || parts[0].startsWith('http')) {
        // Probably a URL.
        return false
    }
    // TODO: we can do further validation here.
    return true
}
