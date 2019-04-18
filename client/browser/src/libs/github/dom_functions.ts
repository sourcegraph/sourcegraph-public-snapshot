import { DiffPart, DOMFunctions } from '@sourcegraph/codeintellify'
import { isDomSplitDiff } from './util'

const getDiffCodePart = (codeElement: HTMLElement): DiffPart => {
    const td = codeElement.closest('td')!

    if (td.classList.contains('blob-code-addition')) {
        return 'head'
    }

    if (td.classList.contains('blob-code-deletion')) {
        return 'base'
    }
    if (isDomSplitDiff()) {
        // If there are more cells on the right, this is the base, otherwise the head
        return td.nextElementSibling ? 'base' : 'head'
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
        // Old versions of GitHub write a +, -, or space character directly into
        // the HTML text of the diff:
        //
        // <span class="blob-code-inner">+	fmt.<span class="pl-c1">Println</span>...
        //                               ^
        //
        // New versions of GitHub do not, and Refined GitHub strips these
        // characters.
        //
        // Since a +, -, or space character in the first column could be either
        //
        // - a diff indicator on an old version of GitHub, or
        // - simply part of the code being diffed on either a new version of
        //   GitHub or Refined GitHub,
        //
        // we check for the presence of other diff indicators that we know are
        // mutually exclusive with the first character diff indicator.

        // Some versions of GitHub have blob-code-marker-* classes instead of the first character diff indicator.
        const blobCodeInner = codeElement.closest('.blob-code-inner')
        const hasBlobCodeMarker =
            blobCodeInner &&
            ['deletion', 'context', 'addition'].some(name =>
                blobCodeInner.classList.contains('blob-code-marker-' + name)
            )

        // Some versions of GitHub have data-code-marker attributes instead of the first character diff indicator.
        const tr = codeElement.closest('tr')
        const hasDataCodeMarkerUnified = tr && tr.querySelector('td[data-code-marker]')
        const hasDataCodeMarkerSplit = blobCodeInner && blobCodeInner.hasAttribute('data-code-marker')
        const hasDataCodeMarker = hasDataCodeMarkerUnified || hasDataCodeMarkerSplit

        // Refined GitHub strips the first character diff indicator.
        const hasRefinedGitHub = codeElement.closest('.refined-github-diff-signs')

        // When no other diff indicator is found, we assume the first character
        // is a diff indicator.
        return !hasBlobCodeMarker && !hasDataCodeMarker && !hasRefinedGitHub
    },
}

/**
 * Implementations of the DOM functions for GitHub blob code views
 */
export const singleFileDOMFunctions: DOMFunctions = {
    getCodeElementFromTarget: getCodeCellFromTarget,
    getCodeElementFromLineNumber: (codeView, line) => {
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
        return codeCell
    },
    getLineNumberFromCodeElement: (codeElement: HTMLElement): number => {
        const cell = codeElement.closest('td')!.previousElementSibling as HTMLTableCellElement

        return parseInt(cell.firstElementChild!.textContent!, 10)
    },
}
