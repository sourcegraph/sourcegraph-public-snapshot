import { DiffPart, DOMFunctions } from '@sourcegraph/codeintellify'
import { isDomSplitDiff } from './util'

const getDiffCodePart = (codeElement: HTMLElement): DiffPart => {
    const td = codeElement.closest('td')!
    if (isDomSplitDiff()) {
        // If there are more cells on the right, this is the base, otherwise the head
        return td.nextElementSibling ? 'base' : 'head'
    }

    if (td.classList.contains('blob-code-addition')) {
        return 'head'
    }

    if (td.classList.contains('blob-code-deletion')) {
        return 'base'
    }

    return 'head'
}

/**
 * Returns the 0-based index of the cell that holds the line number for a given part,
 * depending on whether the diff is in unified or split view.
 * Prefers head.
 */
const getLineNumberElementIndex = (part: DiffPart): number => {
    if (isDomSplitDiff()) {
        return part === 'base' ? 0 : 2
    }
    return part === 'base' ? 0 : 1
}

/**
 * Gets the line number for a given code element on unified diff, split diff and blob views
 */
const getLineNumberFromCodeElement = (codeElement: HTMLElement): number => {
    // In diff views, the code element is the `<span>` inside the cell
    // On blob views, the code element is the `<td>` itself, so `closest()` will simply return it
    // Walk all previous sibling cells until we find one with the line number
    let cell = codeElement.closest('td')!.previousElementSibling as HTMLTableCellElement
    while (cell) {
        if (cell.dataset.lineNumber) {
            return parseInt(cell.dataset.lineNumber, 10)
        }
        cell = cell.previousElementSibling as HTMLTableCellElement
    }
    throw new Error('Could not find a line number in any cell')
}

/**
 * Gets the `<td>` element for a target that contains the code
 */
const getCodeCellFromTarget = (target: HTMLElement): HTMLTableCellElement | null => {
    const cell = target.closest('td.blob-code') as HTMLTableCellElement
    // Handle rows with the [ ↕ ] button that expands collapsed unchanged lines
    if (!cell || cell.parentElement!.classList.contains('js-expandable-line')) {
        return null
    }
    return cell
}

/**
 * Returns the `<span class="blob-code-inner">` element inside a cell.
 * The code element on diff pages is the `<span class="blob-code-inner">` element inside the cell,
 * because the cell also contains a button to add a comment
 */
const getBlobCodeInner = (codeCell: HTMLTableCellElement) =>
    codeCell.classList.contains('blob-code-inner')
        ? codeCell // `<td>`'s in sections of the table that were expanded are not commentable so the `.blob-code-inner` element is the `<td>`
        : (codeCell.querySelector('.blob-code-inner') as HTMLElement)

/**
 * Implementations of the DOM functions for GitHub diff code views
 */
export const diffDomFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => {
        const codeCell = getCodeCellFromTarget(target)
        return codeCell && getBlobCodeInner(codeCell)
    },
    getCodeElementFromLineNumber: (codeView, line, part) => {
        const nthChild = getLineNumberElementIndex(part!) + 1 // nth-child() is 1-indexed
        const lineNumberCell = codeView.querySelector<HTMLTableCellElement>(
            `td:nth-child(${nthChild})[data-line-number="${line}"]`
        )
        if (!lineNumberCell) {
            return null
        }
        let codeCell: HTMLTableCellElement
        if (isDomSplitDiff()) {
            // In split diff view, the code cell is next to the line number cell
            codeCell = lineNumberCell.nextElementSibling as HTMLTableCellElement
        } else {
            // In unified diff view, the code cell is the last cell
            const row = lineNumberCell.parentElement!
            codeCell = row.lastElementChild as HTMLTableCellElement
        }
        return getBlobCodeInner(codeCell)
    },
    getLineNumberFromCodeElement,
    getDiffCodePart,
    isFirstCharacterDiffIndicator: codeElement => {
        // Look for standard GitHub pseudo elements with diff indicators
        const blobCodeInner = codeElement.closest('.blob-code-inner')
        if (!blobCodeInner) {
            throw new Error('Could not find .blob-code-inner element for codeElement')
        }
        if (
            ['deletion', 'context', 'addition'].some(name =>
                blobCodeInner.classList.contains('blob-code-marker-' + name)
            )
        ) {
            return false
        }
        // Old GitHub Enterprise, check for Refined GitHub
        return !codeElement.closest('.refined-github-diff-signs')
    },
}

/**
 * Implementations of the DOM functions for GitHub blob code views
 */
export const blobDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: getCodeCellFromTarget,
    getCodeElementFromLineNumber: (codeView, line, part) => {
        const lineNumberCell = codeView.querySelector(`td[data-line-number="${line}"]`)
        if (!lineNumberCell) {
            return null
        }
        const codeCell = lineNumberCell.nextElementSibling as HTMLTableCellElement
        // In blob views, the `<td>` is the code element
        return codeCell
    },
    getLineNumberFromCodeElement,
}

export const searchCodeSnippetDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: getCodeCellFromTarget,
    getCodeElementFromLineNumber: (codeView, line, part) => {
        const lineNumberCells = codeView.querySelectorAll('td.blob-num')
        let lineNumberCell: HTMLElement | null = null

        for (const cell of lineNumberCells) {
            const a = cell.querySelector('a')!
            if (a.href.match(new RegExp(`#L${line}$`))) {
                lineNumberCell = cell as HTMLElement
                break
            }
        }

        if (!lineNumberCell) {
            return null
        }

        const codeCell = lineNumberCell.nextElementSibling as HTMLTableCellElement
        // In blob views, the `<td>` is the code element
        console.log(codeCell)
        return codeCell
    },
    getLineNumberFromCodeElement: (codeElement: HTMLElement): number => {
        const cell = codeElement.closest('td')!.previousElementSibling as HTMLTableCellElement

        return parseInt(cell.firstElementChild!.textContent!, 10)
    },
}
