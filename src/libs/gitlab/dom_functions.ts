import { DOMFunctions } from '@sourcegraph/codeintellify'
import { CodeView } from '../code_intelligence'

export const singleFileDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('span.line') as HTMLElement | null,
    getLineNumberFromCodeElement: codeElement => {
        const line = codeElement.id.replace(/^LC/, '')
        return parseInt(line, 10)
    },
    getCodeElementFromLineNumber: (codeView, line) => codeView.querySelector<HTMLElement>(`#LC${line}`),
}

const getDiffCodePart: DOMFunctions['getDiffCodePart'] = codeElement => {
    let selector = 'old'

    const row = codeElement.closest('td')!

    // Split diff
    if (row.classList.contains('parallel')) {
        selector = 'left-side'
    }

    return row.classList.contains(selector) ? 'base' : 'head'
}

export const diffDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: singleFileDOMFunctions.getCodeElementFromTarget,
    getLineNumberFromCodeElement: codeElement => {
        const part = getDiffCodePart(codeElement)

        let cell: HTMLElement | null = codeElement.closest('td')
        while (
            cell &&
            !cell.matches(`.diff-line-num.${part === 'base' ? 'old_line' : 'new_line'}`) &&
            cell.previousElementSibling
        ) {
            cell = cell.previousElementSibling as HTMLElement | null
        }

        if (cell) {
            const a = cell.querySelector<HTMLElement>('a')!
            return parseInt(a.dataset.linenumber || '', 10)
        }

        throw new Error('Unable to determine line number for diff code element')
    },
    getCodeElementFromLineNumber: (codeView, line, part) => {
        const lineNumberElement = codeView.querySelector(
            `.${part === 'base' ? 'old_line' : 'new_line'} [data-linenumber="${line}"]`
        )
        if (!lineNumberElement) {
            return null
        }

        const row = lineNumberElement.closest('tr')
        if (!row) {
            return null
        }

        let selector = 'span.line'

        // Split diff
        if (row.classList.contains('parallel')) {
            selector = `.${part === 'base' ? 'left-side' : 'right-side'} ${selector}`
        }

        return row.querySelector<HTMLElement>(selector)
    },
    getDiffCodePart,
}

export const singleFileGetLineRanges: CodeView['getLineRanges'] = codeView => {
    const firstLine = codeView.querySelector<HTMLTableRowElement>('.line:first-of-type')
    if (!firstLine) {
        throw new Error('Unable to determine start line of code view')
    }

    const lastLine = codeView.querySelector<HTMLTableRowElement>('.line:last-of-type')
    if (!lastLine) {
        throw new Error('Unable to determine start line of code view')
    }

    const getLineNumber = (line: HTMLElement) => {
        const codeElement = singleFileDOMFunctions.getCodeElementFromTarget(line)!

        return singleFileDOMFunctions.getLineNumberFromCodeElement(codeElement)
    }

    return [
        {
            start: getLineNumber(firstLine),
            end: getLineNumber(lastLine),
        },
    ]
}

export const diffFileGetLineRanges: CodeView['getLineRanges'] = (codeView, part) => {
    const ranges: { start: number; end: number }[] = []

    let start: number | null = null
    let end: number | null = null

    for (const row of codeView.querySelectorAll<HTMLTableRowElement>('tr')) {
        const isCode = row.firstElementChild && !row.firstElementChild.classList.contains('js-unfold')

        if (isCode) {
            const line = row.querySelector<HTMLElement>(`td.${part === 'base' ? 'left-side' : 'right-side'} span.line`)!
            if (!line) {
                // Empty row
                continue
            }

            const codeElement = diffDOMFunctions.getCodeElementFromTarget(line)!

            const num = diffDOMFunctions.getLineNumberFromCodeElement(codeElement)
            if (start === null) {
                start = num
            } else {
                end = num
            }
        } else if (start && end) {
            ranges.push({ start, end })
        }

        if (!isCode) {
            start = null
            end = null
        }
    }

    return ranges
}
