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
    decorationStyleForTheme,
} from '@sourcegraph/shared/src/api/extension/api/decorations'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import styles from './ColumnDecorator.module.scss'

export interface LineDecoratorProps extends ThemeProps {
    extensionID: string
    decorations: DecorationMapByLine
    codeViewElements: ReplaySubject<HTMLElement | null>
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

            const removeAddedCells = (): void => {
                for (const [cell] of addedCells) {
                    cell.remove()
                }
            }

            const subscription = codeViewElements.subscribe(codeView => {
                if (codeView) {
                    const table = codeView.firstElementChild as HTMLTableElement

                    for (let index = 0; index < table.rows.length; index++) {
                        const row = table.rows[index]
                        const className = extensionID.replace(/\//g, '-')

                        const cell = row.querySelector<HTMLTableCellElement>(`td.${className}`) || row.insertCell(0)
                        cell.classList.add('decoration', className)
                        cell.dataset.lineDecorationAttachmentPortal = 'true'

                        const currentLineDecorations = decorations.get(index + 1)

                        // TODO: do we need this for cells with decorations or all cells?
                        for (const decoration of currentLineDecorations || []) {
                            const style = decorationStyleForTheme(decoration, isLightTheme)

                            if (style.backgroundColor) {
                                cell.style.backgroundColor = style.backgroundColor
                            }
                            if (style.border) {
                                cell.style.border = style.border
                            }
                            if (style.borderColor) {
                                cell.style.borderColor = style.borderColor
                            }
                            if (style.borderWidth) {
                                cell.style.borderWidth = style.borderWidth
                            }
                        }

                        // store created cells
                        addedCells.set(cell, currentLineDecorations)
                    }

                    setPortalNodes(addedCells)
                } else {
                    // code view ref passed `null`, so element is leaving DOM
                    removeAddedCells()
                }
            })

            return () => {
                subscription.unsubscribe()
                removeAddedCells()
            }
        }, [codeViewElements, decorations, isLightTheme, extensionID])

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
                                    <LinkOrSpan
                                        key={`${decoration.after.contentText ?? decoration.after.hoverMessage ?? ''}-${
                                            portalRoot.dataset.line ?? ''
                                        }`}
                                        data-line-decoration-attachment={true}
                                        to={attachment.linkURL}
                                        data-tooltip={attachment.hoverMessage}
                                        // Use target to open external URLs
                                        target={
                                            attachment.linkURL && isAbsoluteUrl(attachment.linkURL)
                                                ? '_blank'
                                                : undefined
                                        }
                                        // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
                                        rel="noreferrer noopener"
                                    >
                                        <span
                                            className={styles.contents}
                                            data-line-decoration-attachment-content={true}
                                            // eslint-disable-next-line react/forbid-dom-props
                                            style={{
                                                color: style.color,
                                                backgroundColor: style.backgroundColor,
                                            }}
                                            data-contents={attachment.contentText || ''}
                                        />
                                    </LinkOrSpan>
                                )
                            }),
                            portalRoot
                        )
                    )
                    .toArray()}
            </>
        )
    }
)

ColumnDecorator.displayName = 'ColumnDecorator'
