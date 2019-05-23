import { DOMFunctions } from '@sourcegraph/codeintellify'

const getLineNumberCell = (codeElement: HTMLElement) => {
    let elem: HTMLElement | null = codeElement
    while ((elem && elem.tagName !== 'TH') || (elem && !elem.textContent)) {
        elem = elem.previousElementSibling as HTMLElement | null
    }
    return elem
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
            (td.classList.contains('show-more') || td.classList.contains('show-context') || !getLineNumberCell(td))
        ) {
            return null
        }

        return td
    },
    getCodeElementFromLineNumber: (codeView, line, part) => {
        const lineNumberCells = codeView.querySelectorAll(`th:nth-of-type(${part === 'base' ? 1 : 2})`)
        for (const lineNumberCell of lineNumberCells) {
            if (lineNumberCell.textContent && parseInt(lineNumberCell.textContent, 10) === line) {
                let codeElement = lineNumberCell as HTMLElement | null
                while (codeElement && (codeElement.tagName !== 'TD' || codeElement.classList.contains('copy'))) {
                    codeElement = codeElement.nextElementSibling as HTMLElement | null
                }

                return codeElement
            }
        }

        return null
    },
    getLineNumberFromCodeElement: codeElement => {
        const elem = getLineNumberCell(codeElement)

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

export const diffusionDOMFns: DOMFunctions = {
    getCodeElementFromTarget: target => target.closest('td'),
    getCodeElementFromLineNumber: (codeView, line) => {
        const row = codeView.querySelector(`tr:nth-of-type(${line})`)
        if (!row) {
            throw new Error(`unable to find row ${line} from code view`)
        }

        return row.querySelector<HTMLElement>('td')
    },
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
        return parseInt(lineAnchor.textContent || '', 10)
    },
    isFirstCharacterDiffIndicator: () => false,
}
