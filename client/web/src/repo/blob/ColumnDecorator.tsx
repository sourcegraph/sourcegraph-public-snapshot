import React, { useLayoutEffect, useCallback, useEffect } from 'react'

import isAbsoluteUrl from 'is-absolute-url'
import iterate from 'iterare'
import { toNumber } from 'lodash'
import ReactDOM from 'react-dom'
import { BehaviorSubject, ReplaySubject } from 'rxjs'

import { isDefined, property } from '@sourcegraph/common'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import {
    decorationAttachmentStyleForTheme,
    DecorationMapByLine,
} from '@sourcegraph/shared/src/api/extension/api/decorations'
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

import styles from './ColumnDecorator.module.scss'

export interface LineDecoratorProps extends ThemeProps {
    extensionID: string
    decorations: DecorationMapByLine
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
export const ColumnDecorator = React.memo<LineDecoratorProps>(
    ({ decorations, isLightTheme, codeViewElements, extensionID }) => {
        const [portalNodes, setPortalNodes] = React.useState<
            Map<HTMLTableCellElement, TextDocumentDecoration[] | undefined>
        >()

        // `ColumnDecorator` uses `useLayoutEffect` instead of `useEffect` in order to synchronously re-render
        // after mount/decoration updates, but before the browser has painted DOM updates.
        // This prevents users from seeing inconsistent states where changes handled by React have been
        // painted, but DOM manipulation handled by these effects are painted on the next tick.
        useLayoutEffect(() => {
            const addedCells = new Map<HTMLTableCellElement, TextDocumentDecoration[] | undefined>()

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
                setPortalNodes(undefined)
            }

            const subscription = codeViewElements.subscribe(codeView => {
                if (codeView) {
                    const table = codeView.firstElementChild as HTMLTableElement

                    for (let index = 0; index < table.rows.length; index++) {
                        const row = table.rows[index]
                        const className = extensionID.replace(/\//g, '-')

                        let cell = row.querySelector<HTMLTableCellElement>(`td.${className}`)
                        if (!cell) {
                            cell = row.insertCell(1)
                            cell.classList.add(styles.decoration, className)

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

                        const currentLineDecorations = decorations.get(index + 1)

                        // store created cells
                        addedCells.set(cell, currentLineDecorations)
                    }

                    setPortalNodes(addedCells)
                } else {
                    // code view ref passed `null`, so element is leaving DOM
                    cleanup()
                }
            })

            return () => {
                subscription.unsubscribe()
                cleanup()
            }
        }, [codeViewElements, decorations, extensionID])

        if (!portalNodes?.size) {
            return null
        }

        return (
            <>
                {iterate(portalNodes)
                    .map(([portalRoot, lineDecorations]) =>
                        ReactDOM.createPortal(
                            <ColumnDecoratorContents
                                line={toNumber(portalRoot.previousElementSibling?.dataset.line)} // TODO: use more stable way to get line number
                                lineDecorations={lineDecorations}
                                isLightTheme={isLightTheme}
                                onSelect={selectRow}
                                onDeselect={deselectRow}
                            />,
                            portalRoot.querySelector(`.${styles.wrapper}`) as HTMLDivElement
                        )
                    )
                    .toArray()}
            </>
        )
    }
)

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

export const ColumnDecoratorContents: React.FunctionComponent<{
    line?: number
    lineDecorations: TextDocumentDecoration[] | undefined
    isLightTheme: boolean
    onSelect?: (line: number) => void
    onDeselect?: (line: number) => void
}> = ({ line, lineDecorations, isLightTheme, onSelect, onDeselect }) => {
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

    return (
        <>
            {lineDecorations?.filter(property('after', isDefined)).map(decoration => {
                const attachment = decoration.after
                const style = decorationAttachmentStyleForTheme(attachment, isLightTheme)

                return (
                    <Popover isOpen={isOpen} onOpenChange={onPopoverOpenChange} key={id}>
                        <PopoverTrigger
                            as={Link}
                            style={{ color: style.color }}
                            to={attachment.linkURL!}
                            // Use target to open external URLs
                            target={attachment.linkURL && isAbsoluteUrl(attachment.linkURL) ? '_blank' : undefined}
                            // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
                            rel="noreferrer noopener"
                            onFocus={open}
                            onBlur={close}
                            onMouseEnter={open}
                            onMouseLeave={closeWithTimeout}
                        >
                            <span
                                className={styles.contents}
                                data-line-decoration-attachment-content={true}
                                data-contents={attachment.contentText || ''}
                            />
                        </PopoverTrigger>

                        <PopoverContent
                            targetPadding={createRectangle(0, 0, 10, 10)}
                            position={Position.top}
                            focusLocked={false}
                            onMouseEnter={resetCloseTimeout}
                            onMouseLeave={close}
                        >
                            {attachment.hoverMessage}
                        </PopoverContent>
                    </Popover>
                )
            })}
        </>
    )
}

ColumnDecorator.displayName = 'ColumnDecorator'
