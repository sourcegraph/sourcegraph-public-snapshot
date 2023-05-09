import { MarkdownLine, parseMarkdown } from '../chat/markdown'

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
    filesExist: (filePaths: string[]) => Promise<{ [filePath: string]: boolean }>
): Promise<HighlightTokensResult> {
    const markdownLines = parseMarkdown(text)
    const tokens = await detectTokens(markdownLines, filesExist)
    const highlightedText = markdownLines
        .map(({ line, isCodeBlock }) => (isCodeBlock ? line : highlightLine(line, tokens)))
        .join('\n')
    return { text: highlightedText, tokens }
}

async function detectTokens(
    lines: MarkdownLine[],
    filesExist: (filePaths: string[]) => Promise<{ [filePath: string]: boolean }>
): Promise<HighlightedToken[]> {
    // mapping from file path to full match
    const filePathToFullMatch: { [filePath: string]: Set<string> } = {}
    for (const { line, isCodeBlock } of lines) {
        if (isCodeBlock) {
            continue
        }
        for (const { fullMatch, pathMatch } of findFilePaths(line)) {
            if (!filePathToFullMatch[pathMatch]) {
                filePathToFullMatch[pathMatch] = new Set<string>()
            }
            filePathToFullMatch[pathMatch].add(fullMatch)
        }
    }

    const filePathsExist = await filesExist([...Object.keys(filePathToFullMatch)])
    const tokens: HighlightedToken[] = []
    for (const [filePath, fullMatches] of Object.entries(filePathToFullMatch)) {
        const exists = filePathsExist[filePath]
        for (const fullMatch of fullMatches) {
            tokens.push({
                type: 'file',
                outerValue: fullMatch,
                innerValue: filePath,
                isHallucinated: !exists,
            })
        }
    }
    return tokens
}

function highlightLine(line: string, tokens: HighlightedToken[]): string {
    let highlightedLine = line
    for (const token of tokens) {
        highlightedLine = highlightedLine.replaceAll(token.outerValue, getHighlightedTokenHTML(token))
    }
    return highlightedLine
}

function getHighlightedTokenHTML(token: HighlightedToken): string {
    const isHallucinatedClassName = token.isHallucinated ? 'hallucinated' : 'not-hallucinated'
    return ` <span class="token-${token.type} token-${isHallucinatedClassName}">${token.outerValue.trim()}</span> `
}

export function findFilePaths(line: string): { fullMatch: string; pathMatch: string }[] {
    const matches: { fullMatch: string; pathMatch: string }[] = []
    for (const m of line.matchAll(filePathRegexp)) {
        const fullMatch = m[0]
        const pathMatch = m[1]
        if (isFilePathLike(fullMatch, pathMatch)) {
            matches.push({ fullMatch, pathMatch })
        }
    }
    return matches
}

const filePathCharacters = '[\\*\\w\\/\\._-]'

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
    if (pathMatch.includes('*')) {
        // Probably a glob pattern
        return false
    }

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
