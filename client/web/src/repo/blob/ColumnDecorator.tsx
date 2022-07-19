import React, { useLayoutEffect } from 'react'

import isAbsoluteUrl from 'is-absolute-url'
import iterate from 'iterare'
import ReactDOM from 'react-dom'
import { ReplaySubject } from 'rxjs'

import { isDefined, property } from '@sourcegraph/common'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import {
    decorationAttachmentStyleForTheme,
    DecorationMapByLine,
} from '@sourcegraph/shared/src/api/extension/api/decorations'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Tooltip } from '@sourcegraph/wildcard'

import styles from './ColumnDecorator.module.scss'

export interface LineDecoratorProps extends ThemeProps {
    extensionID: string
    decorations: DecorationMapByLine
    codeViewElements: ReplaySubject<HTMLElement | null>
}

const selectRow = (event: React.FocusEvent | React.MouseEvent): void => {
    if (event.target instanceof HTMLElement) {
        event.target.closest('tr')?.classList.add('highlighted')
    }
}

const deselectRow = (event: React.FocusEvent | React.MouseEvent): void => {
    if (event.target instanceof HTMLElement) {
        event.target.closest('tr')?.classList.remove('highlighted')
    }
}

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
                            lineDecorations?.filter(property('after', isDefined)).map(decoration => {
                                const attachment = decoration.after
                                const style = decorationAttachmentStyleForTheme(attachment, isLightTheme)

                                return (
                                    <Tooltip
                                        content={attachment.hoverMessage}
                                        key={`${decoration.after.contentText ?? decoration.after.hoverMessage ?? ''}-${
                                            portalRoot.dataset.line ?? ''
                                        }`}
                                    >
                                        <LinkOrSpan
                                            className={styles.item}
                                            // eslint-disable-next-line react/forbid-dom-props
                                            style={{ color: style.color }}
                                            to={attachment.linkURL}
                                            // Use target to open external URLs
                                            target={
                                                attachment.linkURL && isAbsoluteUrl(attachment.linkURL)
                                                    ? '_blank'
                                                    : undefined
                                            }
                                            // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
                                            rel="noreferrer noopener"
                                            onMouseEnter={selectRow}
                                            onMouseLeave={deselectRow}
                                            onFocus={selectRow}
                                            onBlur={deselectRow}
                                        >
                                            <span
                                                className={styles.contents}
                                                data-line-decoration-attachment-content={true}
                                                data-contents={attachment.contentText || ''}
                                            />
                                        </LinkOrSpan>
                                    </Tooltip>
                                )
                            }),
                            portalRoot.querySelector(`.${styles.wrapper}`) as HTMLDivElement
                        )
                    )
                    .toArray()}
            </>
        )
    }
)

ColumnDecorator.displayName = 'ColumnDecorator'
