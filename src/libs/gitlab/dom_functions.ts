import { DOMFunctions } from '@sourcegraph/codeintellify'

export const singleFileDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('span.line') as HTMLElement | null,
    getLineNumberFromCodeElement: codeElement => {
        const line = codeElement.id.replace(/^LC/, '')
        return parseInt(line, 10)
    },
    getCodeElementFromLineNumber: (codeView, line) => codeView.querySelector<HTMLElement>(`#LC${line}`),
}

export const diffDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: singleFileDOMFunctions.getCodeElementFromTarget,
    getLineNumberFromCodeElement: codeElement => {
        let cell: HTMLElement | null = codeElement.closest('td')
        while (cell && !cell.dataset.linenumber && cell.previousElementSibling) {
            cell = cell.previousElementSibling as HTMLElement | null
        }

        if (cell) {
            return parseInt(cell.dataset.linenumber || '', 10)
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
    getDiffCodePart: codeElement => {
        let selector = 'old'

        const row = codeElement.closest('td')!

        // Split diff
        if (row.classList.contains('parallel')) {
            selector = 'left-side'
        }

        return row.classList.contains(selector) ? 'base' : 'head'
    },
}
