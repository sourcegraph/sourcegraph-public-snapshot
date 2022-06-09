import { DiffPart, DOMFunctions } from '@sourcegraph/codeintellify'

import { DiffHunkLineType } from '../../graphql-operations'

export const diffDomFunctions: DOMFunctions = {
    getCodeElementFromTarget: (target: HTMLElement): HTMLTableCellElement | null => {
        const row = target.closest('td')
        if (
            !row ||
            row.getAttribute('data-diff-boundary-content') ||
            row.getAttribute('data-diff-boundary-num') ||
            row.getAttribute('data-hunk-content-empty') ||
            row.getAttribute('data-hunk-num')
        ) {
            return null
        }

        return row
    },

    getLineNumberFromCodeElement: (codeCell: HTMLElement): number => {
        const closestTd = codeCell.closest('td')
        if (closestTd) {
            // First search right-to-left for a line number using `previousElementSibling`
            // If not found, search left-to-right for a line number using `nextElementSibling`
            const siblings = ['previousElementSibling', 'nextElementSibling'] as const
            for (const sibling of siblings) {
                let cell = closestTd[sibling] as HTMLTableCellElement
                while (cell) {
                    if (cell.dataset.line) {
                        return parseInt(cell.dataset.line, 10)
                    }
                    cell = cell[sibling] as HTMLTableCellElement
                }
            }
        }
        throw new Error('Could not find a line number in any cell')
    },

    getDiffCodePart: (codeElement: HTMLElement): DiffPart => {
        const tableCell = codeElement.closest('td') as HTMLTableCellElement
        const tableRow = codeElement.parentElement as HTMLTableRowElement
        const isSplitMode = tableRow.getAttribute('data-diff-mode') === 'split'
        const lineKind = tableRow.getAttribute('data-hunk-line-kind')

        if (lineKind === DiffHunkLineType.ADDED) {
            return 'head'
        }

        if (lineKind === DiffHunkLineType.DELETED) {
            return 'base'
        }

        if (isSplitMode) {
            return tableCell.nextElementSibling ? 'base' : 'head'
        }

        return 'head'
    },

    getCodeElementFromLineNumber: (
        codeView: HTMLElement,
        line: number,
        part?: DiffPart
    ): HTMLTableCellElement | null => {
        // For unchanged lines, prefer line number in head
        const lineNumberCell = codeView.querySelector(
            `[data-line="${line}"][data-part="${part || 'head'}"]`
        ) as HTMLTableCellElement
        if (!lineNumberCell) {
            return null
        }

        const row = lineNumberCell.parentElement as HTMLTableRowElement
        // row.cells.length === 4 is the number of cells for side by side diff
        const codeCell = row.cells.length === 4 ? row.cells[lineNumberCell.cellIndex + 1] : row.cells[2]
        return codeCell
    },

    isFirstCharacterDiffIndicator: () => false,
}
