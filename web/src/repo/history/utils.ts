import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import { escape } from 'lodash'
import marked from 'marked'
import { MarkedString, MarkupContent, MarkupKind } from 'vscode-languageserver-types'
import { ParsedRepoURI } from '..'
import { HoverMerged } from '../../backend/features'
import { displayRepoPath } from '../../components/RepoFileLink'

/** The data stored for each history entry. */
export interface SymbolHistoryEntry {
    /** The symbol name */
    name: string
    repoPath: string
    filePath: string
    url: string
    lineNumber?: number
    /** Hover contents, excluding the symbol name */
    hoverContents: string[]
    /** The actual line of code the symbol is in and 5 surrounding lines */
    linesOfCode: string[]
    /** Combination of name, hover contents, lines of code. */
    rawString: string
}

const highlightCodeSafe = (code: string, language?: string): string => {
    try {
        if (language) {
            return highlight(language, code, true).value
        }
        return highlightAuto(code).value
    } catch (err) {
        console.warn('Error syntax-highlighting hover markdown code block', err)
        return escape(code)
    }
}

export const hoverContentsToString = (contents: (MarkupContent | MarkedString)[]): string[] => {
    const contentList = []
    const hoverContents = contents
    for (let content of hoverContents) {
        let signature: string
        if (typeof content === 'string') {
            const hold = content
            content = { kind: MarkupKind.Markdown, value: hold }
        }
        if (MarkupContent.is(content)) {
            if (content.kind === MarkupKind.Markdown) {
                try {
                    const rendered = marked(content.value, {
                        gfm: true,
                        breaks: true,
                        sanitize: true,
                        highlight: (code, language) => '<code>' + highlightCodeSafe(code, language) + '</code>',
                    })
                    signature = rendered
                } catch (err) {
                    signature = 'errored'
                }
            } else {
                signature = content.value
            }
        } else {
            signature = highlightCodeSafe(content.value, content.language)
        }
        contentList.push(signature)
    }

    return contentList
}

export const getSymbolSignature = (contents: (MarkupContent | MarkedString)[]): string => {
    const symbolSignature = [contents[0]]
    return hoverContentsToString(symbolSignature)[0]
}

export const createSymbolHistoryEntry = (
    parsedRepoURI: ParsedRepoURI,
    hover: HoverMerged,
    fileLines: string[],
    url: string
): SymbolHistoryEntry => {
    const name = getSymbolSignature(hover.contents)
    let hoverContents = hoverContentsToString(hover.contents).slice(1)
    const position = parsedRepoURI.position
    let lineNumber = 0
    let surroundingLinesOfCode = fileLines
    if (position) {
        lineNumber = position.line
        let startLine = lineNumber - 3
        if (startLine < 0) {
            startLine = 0
        }
        const endLine = lineNumber + 3
        surroundingLinesOfCode = surroundingLinesOfCode.slice(startLine, endLine)
    }

    // Only show first 500 characers of hover documentation
    if (hoverContents.length > 0 && hoverContents[0].length > 500) {
        const div = document.createElement('div')
        div.innerHTML = hoverContents[0].slice(0, 500)
        if (div.lastChild && div.lastChild.textContent) {
            const span = document.createElement('span')
            span.textContent = '...'
            div.lastChild.appendChild(span)
        }
        hoverContents = [div.outerHTML]
    }

    let rawString = name
    rawString = rawString.concat(...surroundingLinesOfCode)
    rawString = rawString.concat(...hoverContents)
    const obj: SymbolHistoryEntry = {
        name,
        url,
        repoPath: displayRepoPath(parsedRepoURI.repoPath),
        // The caller should ensure filePath is defined before calling this method
        filePath: parsedRepoURI.filePath as string,
        hoverContents,
        linesOfCode: surroundingLinesOfCode,
        lineNumber,
        rawString,
    }
    return obj
}
