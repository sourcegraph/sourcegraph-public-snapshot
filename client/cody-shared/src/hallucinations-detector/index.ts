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
    const markdownLines = parseMarkdown(text)
    const tokens = await detectTokens(markdownLines, fileExists)
    const highlightedText = markdownLines
        .map(({ line, isCodeBlock }) => (isCodeBlock ? line : highlightLine(line, tokens)))
        .join('\n')
    return { text: highlightedText, tokens }
}

async function detectTokens(
    lines: MarkdownLine[],
    fileExists: (filePath: string) => Promise<boolean>
): Promise<HighlightedToken[]> {
    const tokens: HighlightedToken[] = []
    for (const { line, isCodeBlock } of lines) {
        if (isCodeBlock) {
            continue
        }
        const lineTokens = await detectFilePaths(line, fileExists)
        tokens.push(...lineTokens)
    }
    return deduplicateTokens(tokens)
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

interface MarkdownLine {
    line: string
    isCodeBlock: boolean
}

function parseMarkdown(text: string): MarkdownLine[] {
    const markdownLines: MarkdownLine[] = []
    let isCodeBlock = false
    for (const line of text.split('\n')) {
        if (line.trim().startsWith('```')) {
            markdownLines.push({ line, isCodeBlock: true })
            isCodeBlock = !isCodeBlock
        } else {
            markdownLines.push({ line, isCodeBlock })
        }
    }
    return markdownLines
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
