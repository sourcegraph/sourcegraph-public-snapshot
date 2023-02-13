import React, { useLayoutEffect } from 'react'

import ReactDOM from 'react-dom'
import { useNavigate } from 'react-router-dom-v5-compat'
import { ReplaySubject } from 'rxjs'

import { BlameHunk } from '../blame/useBlameHunks'

import { BlameDecoration } from './BlameDecoration'

import styles from './BlameColumn.module.scss'

interface BlameColumnProps {
    isBlameVisible?: boolean
    blameHunks?: { current: BlameHunk[] | undefined; firstCommitDate: Date | undefined }
    isLightTheme: boolean
    codeViewElements: ReplaySubject<HTMLElement | null>
}

const getRowByLine = (line: number): HTMLTableRowElement | null | undefined =>
    [...document.querySelectorAll('table')]
        .find(table => table.querySelector(`.${styles.decoration}`))
        ?.querySelector(`tr:nth-of-type(${line})`)

const selectRow = (line: number): void => getRowByLine(line)?.classList.add('highlighted')
const deselectRow = (line: number): void => getRowByLine(line)?.classList.remove('highlighted')

export const BlameColumn = React.memo<BlameColumnProps>(
    ({ isBlameVisible, codeViewElements, blameHunks, isLightTheme }) => {
        const navigate = useNavigate()
        /**
         * Array to store the DOM element and the blame hunk to render in it.
         * As blame decorations are displayed in the column view, we need to add a corresponding
         * cell to each row regrdless of whether there is a blame hunk to render in it or not (empty cell).
         * Array length equals to the number of rows in the table.
         * Array index represents 0-based line number.
         */
        const [cells, setCells] = React.useState<[HTMLTableCellElement, BlameHunk | undefined][]>([])

        /*
        `BlameColumn` uses `useLayoutEffect` instead of `useEffect` in order to synchronously re-render
        after mount/decoration updates, but before the browser has painted DOM updates.
        This prevents users from seeing inconsistent states where changes handled by React have been
        painted, but DOM manipulation handled by these effects are painted on the next tick.
     */
        useLayoutEffect(() => {
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
                if (codeView?.firstElementChild instanceof HTMLTableElement) {
                    const table = codeView.firstElementChild

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
                            if (isBlameVisible) {
                                // ensure blame column has needed width to avoid content jumping after blame hunks are loaded
                                wrapper.classList.add(styles.visible)
                            }
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

                        const currentLineDecorations = blameHunks?.current?.find(hunk => hunk.startLine - 1 === index)

                        // store created cell and corresponding blame hunk (or undefined if no blame hunk)
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
        }, [codeViewElements, isBlameVisible, blameHunks])

        return (
            <>
                {cells.map(([portalRoot, blameHunk], index) =>
                    ReactDOM.createPortal(
                        <BlameDecoration
                            line={index + 1}
                            blameHunk={blameHunk}
                            navigate={navigate}
                            onSelect={selectRow}
                            onDeselect={deselectRow}
                            firstCommitDate={blameHunks?.firstCommitDate}
                            isLightTheme={isLightTheme}
                            hideRecency={true}
                        />,
                        // The classname can contain a +, so we would either need to escape it (boo!),
                        // or just use getElementsByClassName.
                        // eslint-disable-next-line unicorn/prefer-query-selector
                        portalRoot.getElementsByClassName(styles.wrapper)[0] as HTMLDivElement
                    )
                )}
            </>
        )
    }
)

BlameColumn.displayName = 'BlameColumn'
