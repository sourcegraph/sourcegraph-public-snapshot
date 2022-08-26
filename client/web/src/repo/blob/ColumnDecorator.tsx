import React, { useLayoutEffect, useCallback, useEffect } from 'react'

import ReactDOM from 'react-dom'
import { BehaviorSubject, ReplaySubject } from 'rxjs'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
    createRectangle,
    useObservable,
    Link,
} from '@sourcegraph/wildcard'

import { BlameHunk } from '../blame/useBlameDecorations'

import styles from './ColumnDecorator.module.scss'

export interface LineDecoratorProps extends ThemeProps {
    blameHunks?: BlameHunk[]
    codeViewElements: ReplaySubject<HTMLElement | null>
}

const getRowByLine = (line: number): HTMLTableRowElement | null | undefined =>
    [...document.querySelectorAll('table')]
        .find(table => table.querySelector(`.${styles.decoration}`)) // TODO: use more stable way to the proper code view element
        ?.querySelector(`tr:nth-of-type(${line})`)

const selectRow = (line: number): void => getRowByLine(line)?.classList.add('highlighted')
const deselectRow = (line: number): void => getRowByLine(line)?.classList.remove('highlighted')

/**
 * Component that prepends lines of code with attachments set by extensions
 */
export const ColumnDecorator = React.memo<LineDecoratorProps>(({ isLightTheme, codeViewElements, blameHunks }) => {
    const [cells, setCells] = React.useState<[HTMLTableCellElement, BlameHunk | undefined][]>([])

    // `ColumnDecorator` uses `useLayoutEffect` instead of `useEffect` in order to synchronously re-render
    // after mount/decoration updates, but before the browser has painted DOM updates.
    // This prevents users from seeing inconsistent states where changes handled by React have been
    // painted, but DOM manipulation handled by these effects are painted on the next tick.
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

                    const currentLineDecorations = blameHunks?.find(hunk => hunk.startLine - 1 === index)

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
                    <ColumnDecoratorContents
                        line={index + 1}
                        blameHunk={blameHunk}
                        isLightTheme={isLightTheme}
                        onSelect={selectRow}
                        onDeselect={deselectRow}
                    />,
                    portalRoot.querySelector(`.${styles.wrapper}`) as HTMLDivElement
                )
            )}
        </>
    )
})

const currentPopoverId = new BehaviorSubject<string | null>(null)
let timeoutId: NodeJS.Timeout | null = null
const resetTimeout = (): void => {
    if (timeoutId) {
        clearTimeout(timeoutId)
        timeoutId = null
    }
}

const usePopover = ({
    id,
    timeout,
    onOpen,
    onClose,
}: {
    id: string
    timeout: number
    onOpen?: () => void
    onClose?: () => void
}): {
    isOpen: boolean
    open: () => void
    close: () => void
    closeWithTimeout: () => void
    resetCloseTimeout: () => void
} => {
    const popoverId = useObservable(currentPopoverId)

    const isOpen = popoverId === id
    useEffect(() => {
        if (isOpen) {
            onOpen?.()
        }

        return () => {
            if (isOpen) {
                onClose?.()
            }
        }
    }, [isOpen, onOpen, onClose])

    const open = useCallback(() => currentPopoverId.next(id), [id])

    const close = useCallback(() => {
        if (currentPopoverId.getValue() === id) {
            currentPopoverId.next(null)
        }
    }, [id])

    const closeWithTimeout = useCallback(() => {
        timeoutId = setTimeout(close, timeout)
    }, [close, timeout])

    return { isOpen, open, close, closeWithTimeout, resetCloseTimeout: resetTimeout }
}

const now = Date.now()

export const ColumnDecoratorContents: React.FunctionComponent<{
    line: number
    blameHunk?: BlameHunk
    isLightTheme: boolean
    onSelect?: (line: number) => void
    onDeselect?: (line: number) => void
}> = ({ line, blameHunk, isLightTheme, onSelect, onDeselect }) => {
    const id = line?.toString() || ''
    const onOpen = useCallback(() => {
        if (typeof line === 'number' && onSelect) {
            onSelect(line)
        }
    }, [line, onSelect])
    const onClose = useCallback(() => {
        if (typeof line === 'number' && onDeselect) {
            onDeselect(line)
        }
    }, [line, onDeselect])
    const { isOpen, open, close, closeWithTimeout, resetCloseTimeout } = usePopover({
        id,
        timeout: 1000,
        onOpen,
        onClose,
    })

    const onPopoverOpenChange = useCallback(() => (isOpen ? close() : open()), [isOpen, close, open])

    if (!blameHunk) {
        return null
    }

    return (
        <Popover isOpen={isOpen} onOpenChange={onPopoverOpenChange} key={id}>
            <PopoverTrigger
                as={Link}
                to={blameHunk.displayInfo.linkURL}
                target="_blank"
                rel="noreferrer noopener"
                className={styles.item}
                onFocus={open}
                onBlur={close}
                onMouseEnter={open}
                onMouseLeave={closeWithTimeout}
            >
                <span
                    className={styles.contents}
                    data-line-decoration-attachment-content={true}
                    data-contents={blameHunk.displayInfo.message}
                />
            </PopoverTrigger>

            <PopoverContent
                targetPadding={createRectangle(0, 0, 10, 10)}
                position={Position.top}
                focusLocked={false}
                onMouseEnter={resetCloseTimeout}
                onMouseLeave={close}
            >
                <div>
                    {blameHunk.displayInfo.displayName} {blameHunk.displayInfo.dateString}
                    <hr />
                    {blameHunk.message}
                </div>
            </PopoverContent>
        </Popover>
    )
}

ColumnDecorator.displayName = 'ColumnDecorator'
