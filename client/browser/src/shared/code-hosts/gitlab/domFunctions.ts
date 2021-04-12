import { DiffPart } from '@sourcegraph/codeintellify'

import { DOMFunctions } from '../shared/codeViews'

const getSingleFileCodeElementFromLineNumber = (
    codeView: HTMLElement,
    line: number,
    part?: DiffPart
): HTMLElement | null => codeView.querySelector<HTMLElement>(`#LC${line}`)
export const singleFileDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('span.line'),
    getLineNumberFromCodeElement: codeElement => {
        const line = codeElement.id.replace(/^LC/, '')
        return parseInt(line, 10)
    },
    getCodeElementFromLineNumber: getSingleFileCodeElementFromLineNumber,
    getLineElementFromLineNumber: getSingleFileCodeElementFromLineNumber,
}

const getDiffCodePart: DOMFunctions['getDiffCodePart'] = codeElement => {
    let selector = 'old'

    const row = codeElement.closest('.diff-td,td')!

    // Split diff
    if (row.classList.contains('parallel')) {
        selector = 'left-side'
    }

    return row.classList.contains(selector) ? 'base' : 'head'
}

const getDiffCodeElementFromLineNumber = (codeView: HTMLElement, line: number, part?: DiffPart): HTMLElement | null => {
    const lineNumberElement = codeView.querySelector<HTMLElement>(
        `.${part === 'base' ? 'old_line' : 'new_line'} [data-linenumber="${line}"]`
    )
    if (!lineNumberElement) {
        return null
    }

    const row = lineNumberElement.closest('.diff-tr,tr')
    if (!row) {
        return null
    }

    let selector = 'span.line'

    // Split diff
    if (row.classList.contains('parallel')) {
        selector = `.${part === 'base' ? 'left-side' : 'right-side'} ${selector}`
    }

    return row.querySelector<HTMLElement>(selector)
}

export const diffDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: singleFileDOMFunctions.getCodeElementFromTarget,
    getLineNumberFromCodeElement: codeElement => {
        const part = getDiffCodePart(codeElement)

        let cell: HTMLElement | null = codeElement.closest('.diff-td,td')
        while (
            cell &&
            // It's possible for a line number container to not contain an <a> tag with the line
            // number, e.g. right side 'old_line' for a deleted file
            !(cell.matches(`.diff-line-num.${part === 'base' ? 'old_line' : 'new_line'}`) && cell.querySelector('a')) &&
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
    getCodeElementFromLineNumber: getDiffCodeElementFromLineNumber,
    getLineElementFromLineNumber: (codeView, line, part) => {
        const codeElement = getDiffCodeElementFromLineNumber(codeView, line, part)
        return codeElement && (codeElement.parentElement as HTMLTableCellElement)
    },
    getDiffCodePart,
}
