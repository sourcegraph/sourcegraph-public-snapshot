import React, { useEffect } from 'react'

import ReactDOM from 'react-dom'
import { ReplaySubject } from 'rxjs'

import { BlameHunk } from '../blame/useBlameDecorations'

import { BlameDecoration } from './BlameDecoration'

import styles from './BlameColumn.module.scss'

interface ColumnDecoratorProps {
    blameHunks: BlameHunk[]
    codeViewElements: ReplaySubject<HTMLElement | null>
}

const getRowByLine = (line: number): HTMLTableRowElement | null | undefined =>
    [...document.querySelectorAll('table')]
        .find(table => table.querySelector(`.${styles.decoration}`))
        ?.querySelector(`tr:nth-of-type(${line})`)

const selectRow = (line: number): void => getRowByLine(line)?.classList.add('highlighted')
const deselectRow = (line: number): void => getRowByLine(line)?.classList.remove('highlighted')

export const BlameColumn = React.memo<ColumnDecoratorProps>(({ codeViewElements, blameHunks }) => {
    const [cells, setCells] = React.useState<[HTMLTableCellElement, BlameHunk | undefined][]>([])

    useEffect(() => {
        const addedCells: [HTMLTableCellElement, BlameHunk | undefined][] = []

        const cleanup = (): void => {
            // remove added cells
            for (const [cell] of addedCells) {
                const row = cell.closest('tr')
                cell.remove()

                // if no other columns with decorations
                if (!row?.querySelector(`.${styles.decoration}`)) {
                    // remove line number cell extra horizontal padding
                    row?.querySelector('td.line')?.classList.remove('px-2')
                }
            }

            // reset state
            setCells([])
        }

        const subscription = codeViewElements.subscribe(codeView => {
            if (codeView) {
                const table = codeView.firstElementChild as HTMLTableElement

                for (let index = 0; index < table.rows.length; index++) {
                    const row = table.rows[index]
                    let cell = row.querySelector<HTMLTableCellElement>(`td.${styles.decoration}`)
                    if (!cell) {
                        cell = row.insertCell(1)
                        cell.classList.add(styles.decoration)

                        // add line number cell extra horizontal padding
                        row.querySelector('td.line')?.classList.add('px-2')

                        // add decorations wrapper
                        const wrapper = document.createElement('div')
                        wrapper.classList.add(styles.wrapper)
                        cell.append(wrapper)

                        // add extra spacers to first and last rows (if table has only one row add both spacers)
                        if (index === 0) {
                            const spacer = document.createElement('div')
                            spacer.classList.add('top-spacer')
                            cell.prepend(spacer)
                        }

                        if (index === table.rows.length - 1) {
                            const spacer = document.createElement('div')
                            spacer.classList.add('bottom-spacer')
                            cell.append(spacer)
                        }
                    }

                    const currentLineDecorations = blameHunks.find(hunk => hunk.startLine - 1 === index)

                    // store created cells
                    addedCells.push([cell, currentLineDecorations])
                }

                setCells(addedCells)
            } else {
                // code view ref passed `null`, so element is leaving DOM
                cleanup()
            }
        })

        return () => {
            subscription.unsubscribe()
            cleanup()
        }
    }, [codeViewElements, blameHunks])

    return (
        <>
            {cells.map(([portalRoot, blameHunk], index) =>
                ReactDOM.createPortal(
                    <BlameDecoration
                        line={index + 1}
                        blameHunk={blameHunk}
                        onSelect={selectRow}
                        onDeselect={deselectRow}
                    />,
                    portalRoot.querySelector(`.${styles.wrapper}`) as HTMLDivElement
                )
            )}
        </>
    )
})

BlameColumn.displayName = 'BlameColumn'
