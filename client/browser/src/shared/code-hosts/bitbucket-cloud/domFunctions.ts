import type { DiffPart } from '@sourcegraph/codeintellify'

import type { DOMFunctions } from '../shared/codeViews'

export const singleFileDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('.view-line > span'),
    getLineNumberFromCodeElement: codeElement => {
        // Line elements and line number elements belong to two seperate column
        // elements in Monaco Editor.
        // There's also no data attribute/class/ID on the line element that tells
        // us what line number it's associated with.
        // Consequently, the most reliable way to get line number from line (+ vice versa)
        // is to determine its relative position in the virtualized view and find its
        // counterpart element at that position in the other column.

        const someLineNumber = document.querySelector<HTMLElement>('.line-number')
        if (!someLineNumber) {
            throw new Error('No line number elements found on the currently viewed page')
        }

        const editor = codeElement.closest('.react-monaco-editor-container')
        if (!editor) {
            throw new Error('No editor found')
        }

        const lineElement = codeElement.closest('.view-line')
        // `querySelectorAll` returns nodes in document order:
        // https://www.w3.org/TR/selectors-api/#queryselectorall,
        // so we can align the associated line and line number elements.
        // We have to do this because there's no seemingly stable class or attribute on the
        // line number elements' container (like '.line-numbers').
        const lineElements = editor.querySelectorAll<HTMLElement>('.view-line')

        const lineElementIndex = [...lineElements].findIndex(element => element === lineElement)
        const lineNumberElements = editor.querySelectorAll<HTMLElement>('.line-number')
        const inferredLineNumberElement = lineNumberElements[lineElementIndex]

        if (inferredLineNumberElement) {
            let lineNumber = parseInt(inferredLineNumberElement.dataset.lineNum ?? '', 10)
            if (!isNaN(lineNumber)) {
                return lineNumber
            }

            lineNumber = parseInt(inferredLineNumberElement.textContent?.trim() ?? '', 10)
            if (!isNaN(lineNumber)) {
                return lineNumber
            }
        }

        throw new Error('Could not find line number')
    },
    getLineElementFromLineNumber,
    getCodeElementFromLineNumber: (codeView, line) => {
        const lineElement = getLineElementFromLineNumber(codeView, line)

        if (!lineElement) {
            return null
        }

        const codeElement = lineElement.querySelector<HTMLElement>(':scope > span')
        if (!codeElement) {
            console.error(`Could not find code element inside .view-line container for line #${line}`)
        }

        return codeElement
    },
}

function getLineElementFromLineNumber(codeView: HTMLElement, line: number): HTMLElement | null {
    // Line elements and line number elements belong to two seperate column
    // elements in Monaco Editor.
    // There's also no data attribute/class/ID on the line element that tells
    // us what line number it's associated with.
    // Consequently, the most reliable way to get line number from line (+ vice versa)
    // is to determine its relative position in the virtualized view and find its
    // counterpart element at that position in the other column.

    let lineNumberElement = codeView.querySelector<HTMLElement>(`[data-line-num="${line}"]`)
    if (!lineNumberElement) {
        for (const element of codeView.querySelectorAll<HTMLElement>('.line-number')) {
            const currentLine = parseInt(element.textContent ?? '', 10)
            if (currentLine === line) {
                lineNumberElement = element
                break
            }
        }
    }

    if (!lineNumberElement) {
        console.error(`Could not find line number element for line #${line}`)
        return null
    }

    // `querySelectorAll` returns nodes in document order:
    // https://www.w3.org/TR/selectors-api/#queryselectorall,
    // so we can align the associated line and line number elements.
    // We have to do this because there's no seemingly stable class or attribute on the
    // line number elements' container (like '.line-numbers').
    const lineNumberElements = codeView.querySelectorAll<HTMLElement>('.line-number')
    const lineNumberElementIndex = [...lineNumberElements].findIndex(element => element === lineNumberElement)

    const lineElements = codeView.querySelectorAll<HTMLElement>('.view-line')
    const inferredLineElement = lineElements[lineNumberElementIndex]

    if (!inferredLineElement) {
        console.error(`Could not find line element for line #${line}`)
    }

    return inferredLineElement
}

function getPRLineElementFromLineNumber(codeView: HTMLElement, line: number, part?: DiffPart): HTMLElement {
    const lineElement = codeView
        .querySelector<HTMLElement>(`[aria-label="${part === 'base' ? 'From' : 'To'} line ${line}"]`)
        ?.closest<HTMLElement>('.line-wrapper')
        ?.querySelector<HTMLElement>('.code-component')

    if (!lineElement) {
        throw new Error(`Could not locate line number element for line ${line}, part: ${String(part)}`)
    }

    return lineElement
}

export const pullRequestDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('.code-diff'),
    getLineNumberFromCodeElement: codeElement => {
        const lineWrapper = codeElement.closest<HTMLElement>('.line-wrapper')
        // Same selector for unified + side-by-side diff views
        const lineNumberPermalinks = lineWrapper?.querySelectorAll<HTMLElement>('.line-number-permalink')
        if (!lineNumberPermalinks || lineNumberPermalinks.length === 0) {
            throw new Error('Could not find line number permalink(s) for code element')
        }

        // eslint-disable-next-line unicorn/prefer-at
        const lastLineNumberPermalink = lineNumberPermalinks[lineNumberPermalinks.length - 1]
        const lineNumber = parseInt(lastLineNumberPermalink.textContent ?? '', 10)
        if (!isNaN(lineNumber)) {
            return lineNumber
        }

        throw new Error('Could not find line number for code element')
    },
    getLineElementFromLineNumber: getPRLineElementFromLineNumber,
    getCodeElementFromLineNumber: (codeView, line, part) =>
        getPRLineElementFromLineNumber(codeView, line, part)?.querySelector<HTMLElement>('.code-diff'),
    getDiffCodePart: codeElement => {
        const lineElement = codeElement.closest('.code-component')
        const lineType = lineElement?.querySelector('.line-type')

        if (lineType?.textContent?.includes('+')) {
            return 'head'
        }
        if (lineType?.textContent?.includes('+')) {
            return 'base'
        }

        // If lineType is empty and this is the unified view, assume it is head.
        if (!codeElement.closest('.side-by-side')) {
            return 'head'
        }

        const lineWrapper = codeElement.closest('.line-wrapper')
        if (lineWrapper?.id) {
            return lineWrapper?.id.includes('oldline') ? 'base' : 'head'
        }

        // Fallback if ID syntax changes: determine diff part
        // by checking whether the line is on the left or the right side.
        const position = [...(lineWrapper?.parentElement?.childNodes || [])].findIndex(
            element => element === lineWrapper
        )
        return position === 0 ? 'base' : 'head'
    },
    isFirstCharacterDiffIndicator: () => false,
}

function getCommitLineElementFromLineNumber(codeView: HTMLElement, line: number, part?: DiffPart): HTMLElement {
    const lineElement = codeView
        .querySelector<HTMLElement>(`[data-${part === 'base' ? 'f' : 't'}num="${line}"]`)
        ?.closest<HTMLElement>('.udiff-line')

    if (!lineElement) {
        throw new Error(`Could not locate line number element for line ${line}, part: ${String(part)}`)
    }

    return lineElement
}

// Commit view only has unified view. Side-by-side is rendered in a modal.
export const commitDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('.source'),
    getLineNumberFromCodeElement: codeElement => {
        const lineElement = codeElement.closest('.udiff-line')
        if (!lineElement) {
            throw new Error('Could not find line element')
        }
        const lineNumbersElement = lineElement.querySelector<HTMLElement>('.line-numbers')
        if (!lineNumbersElement) {
            throw new Error('Could not find line numbers element')
        }

        const lineNumberTo = lineNumbersElement?.dataset.tnum
        if (lineNumberTo) {
            const lineNumber = parseInt(lineNumberTo, 10)
            if (!isNaN(lineNumber)) {
                return lineNumber
            }
        }

        const lineNumberFrom = lineNumbersElement?.dataset.fnum
        if (lineNumberFrom) {
            const lineNumber = parseInt(lineNumberFrom, 10)
            if (!isNaN(lineNumber)) {
                return lineNumber
            }
        }

        throw new Error('Could not find line number for code element')
    },
    getLineElementFromLineNumber: getCommitLineElementFromLineNumber,
    getCodeElementFromLineNumber: (codeView, line, part) =>
        getCommitLineElementFromLineNumber(codeView, line, part)?.querySelector('.source'),
    getDiffCodePart: codeElement => {
        const lineElement = codeElement.closest<HTMLElement>('.udiff-line')
        if (!lineElement) {
            throw new Error('Could not find line element for code element')
        }

        if (lineElement.classList.contains('deletion')) {
            return 'base'
        }
        // Default for unchanged lines or "additions"
        return 'head'
    },
    isFirstCharacterDiffIndicator: () => true,
}
