import type { DiffPart } from '@sourcegraph/codeintellify'

import type { DOMFunctions } from '../shared/codeViews'

const getSingleFileCodeElementFromLineNumber = (
    codeView: HTMLElement,
    line: number,
    part?: DiffPart
): HTMLElement | null => codeView.querySelector<HTMLElement>(`#LC${line}`)

export const singleFileDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('div.line'),
    getLineNumberFromCodeElement: codeElement => {
        const line = codeElement.id.replace(/^LC/, '')
        return parseInt(line, 10)
    },
    getCodeElementFromLineNumber: getSingleFileCodeElementFromLineNumber,
    getLineElementFromLineNumber: getSingleFileCodeElementFromLineNumber,
}

/**
 * Implementation of `getDiffCodePart` for earlier verisons of GitLab
 */
const getDiffCodePartLegacy: DOMFunctions['getDiffCodePart'] = codeElement => {
    let selector = 'old'

    const row = codeElement.closest('.diff-td,td')!

    // Split diff
    if (row.classList.contains('parallel')) {
        selector = 'left-side'
    }

    return row.classList.contains(selector) ? 'base' : 'head'
}

const getDiffCodePart: DOMFunctions['getDiffCodePart'] = codeElement => {
    const interopParent = codeElement.closest<HTMLElement>('[data-interop-type]')

    if (!interopParent) {
        return getDiffCodePartLegacy(codeElement)
    }

    return interopParent.dataset.interopType === 'old' ? 'base' : 'head'
}

/**
 * Implementation of `getDiffCodeElementFromLineNumber` for earlier verisons of GitLab
 */
const getDiffCodeElementFromLineNumberLegacy = (
    codeView: HTMLElement,
    line: number,
    part?: DiffPart
): HTMLElement | null => {
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

const getDiffCodeElementFromLineNumber = (codeView: HTMLElement, line: number, part?: DiffPart): HTMLElement | null => {
    const type = part === 'base' ? 'old' : 'new'

    const interopChild = codeView.querySelector(`[data-interop-${type}-line="${line}"]`)

    if (!interopChild) {
        return getDiffCodeElementFromLineNumberLegacy(codeView, line, part)
    }

    return interopChild.querySelector('span.line')
}

/**
 * Implementation of `getDiffLineNumberFromCodeElement` for earlier verisons of GitLab
 */
const getDiffLineNumberFromCodeElementLegacy: DOMFunctions['getLineNumberFromCodeElement'] = codeElement => {
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
}

const getDiffLineNumberFromCodeElement: DOMFunctions['getLineNumberFromCodeElement'] = codeElement => {
    const interopParent = codeElement.closest<HTMLElement>('[data-interop-line]')

    if (!interopParent) {
        return getDiffLineNumberFromCodeElementLegacy(codeElement)
    }

    return parseInt(interopParent.dataset.interopLine || '', 10)
}

export const diffDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: singleFileDOMFunctions.getCodeElementFromTarget,
    getLineNumberFromCodeElement: getDiffLineNumberFromCodeElement,
    getCodeElementFromLineNumber: getDiffCodeElementFromLineNumber,
    getLineElementFromLineNumber: (codeView, line, part) => {
        const codeElement = getDiffCodeElementFromLineNumber(codeView, line, part)
        return codeElement && (codeElement.parentElement as HTMLTableCellElement)
    },
    getDiffCodePart,
}
