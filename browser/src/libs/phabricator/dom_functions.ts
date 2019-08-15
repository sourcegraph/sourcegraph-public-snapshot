import { DiffPart } from '@sourcegraph/codeintellify'
import { DOMFunctions } from '../code_intelligence/code_views'

const getLineNumberCellFromCodeElement = (codeElement: HTMLElement): HTMLElement | null => {
    let elem: HTMLElement | null = codeElement
    while ((elem && elem.tagName !== 'TH') || (elem && !elem.textContent)) {
        elem = elem.previousElementSibling as HTMLElement | null
    }
    return elem
}

const getDiffLineNumberElementFromLineNumber = (
    codeView: HTMLElement,
    line: number,
    part?: DiffPart
): HTMLElement | null => {
    const lineNumberCells = codeView.querySelectorAll<HTMLTableHeaderCellElement>(
        `th:nth-of-type(${part === 'base' ? 1 : 2})`
    )
    for (const lineNumberCell of lineNumberCells) {
        if (lineNumberCell.textContent && parseInt(lineNumberCell.textContent, 10) === line) {
            return lineNumberCell
        }
    }
    return null
}

const getDiffCodeElementFromLineNumber = (codeView: HTMLElement, line: number, part?: DiffPart): HTMLElement | null => {
    const lineNumberCell = getDiffLineNumberElementFromLineNumber(codeView, line, part)
    let codeElement: HTMLElement | null = lineNumberCell
    while (codeElement && (codeElement.tagName !== 'TD' || codeElement.classList.contains('copy'))) {
        codeElement = codeElement.nextElementSibling as HTMLElement | null
    }
    return codeElement
}

/**
 * Implementations of the DOM functions for diff code views on Phabricator
 */
export const diffDomFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => {
        if (target.tagName === 'TH' || target.classList.contains('copy')) {
            return null
        }

        const td = target.closest('td')
        if (
            td &&
            (td.classList.contains('show-more') ||
                td.classList.contains('show-context') ||
                !getLineNumberCellFromCodeElement(td))
        ) {
            return null
        }

        return td
    },
    getCodeElementFromLineNumber: getDiffCodeElementFromLineNumber,
    getLineElementFromLineNumber: getDiffCodeElementFromLineNumber,
    getLineNumberFromCodeElement: codeElement => {
        const elem = getLineNumberCellFromCodeElement(codeElement)

        if (elem === null) {
            throw new Error('could not find line number element from code element')
        }

        return parseInt(elem.textContent!, 10)
    },
    getDiffCodePart: codeElement => {
        // Changed lines have handy class names.
        if (codeElement.classList.contains('old')) {
            return 'base'
        }
        if (codeElement.classList.contains('new')) {
            return 'head'
        }

        // For diffs, we'll have to traverse back to the line number <th> and see if it is the last element to determin
        // whether it was the base or head.
        let elem: HTMLElement = codeElement
        while (elem.tagName !== 'TH') {
            if (!elem.previousElementSibling) {
                throw Error('could not find line number cell from code element')
            }
            elem = elem.previousElementSibling as HTMLElement
        }

        // In unified diffs, both <th>'s have a class telling us which side of the diff the line belongs to.
        if (elem.classList.contains('left')) {
            return 'base'
        }
        if (elem.classList.contains('right')) {
            return 'head'
        }

        return elem.previousElementSibling ? 'head' : 'base'
    },
    isFirstCharacterDiffIndicator: (codeElement: HTMLElement) => {
        const firstChild = codeElement.firstElementChild as HTMLElement
        if (firstChild.classList.contains('aural-only')) {
            return true
        }

        return false
    },
}

const getDiffusionCodeElementFromLineNumber = (
    codeView: HTMLElement,
    line: number,
    part?: DiffPart
): HTMLElement | null => {
    const row = codeView.querySelector<HTMLTableRowElement>(`tr:nth-of-type(${line})`)
    if (!row) {
        throw new Error(`unable to find row ${line} from code view`)
    }
    return row.querySelector('td')
}

export const diffusionDOMFns: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('td'),
    getCodeElementFromLineNumber: getDiffusionCodeElementFromLineNumber,
    getLineElementFromLineNumber: getDiffusionCodeElementFromLineNumber,
    getLineNumberFromCodeElement: codeElement => {
        let lineCell = codeElement as HTMLElement | null
        while (
            lineCell !== null &&
            lineCell.tagName !== 'TH' &&
            !lineCell.classList.contains('phabricator-source-line')
        ) {
            lineCell = lineCell.previousElementSibling as HTMLElement | null
        }
        if (!lineCell) {
            throw new Error('could not find line number cell from code element')
        }

        const lineAnchor = lineCell.querySelector('a')
        if (!lineAnchor) {
            throw new Error('could not find line number anchor from code element')
        }
        // In recent Phabricator versions, the line number is stored in the `data-n`
        // attribute, and the textContent is empty.
        if (lineAnchor.dataset.n !== undefined) {
            const lineNumber = parseInt(lineAnchor.dataset.n, 10)
            if (isNaN(lineNumber)) {
                throw new Error(`Could not parse lineNumber from data-n attribute: ${lineAnchor.dataset.n}`)
            }
            return lineNumber
        }
        const lineNumber = parseInt(lineAnchor.textContent || '', 10)
        if (isNaN(lineNumber)) {
            throw new Error(`Could not parse lineNumber from lineAnchor.textContent: ${lineAnchor.textContent}`)
        }
        return lineNumber
    },
    isFirstCharacterDiffIndicator: () => false,
}
