import React, { useLayoutEffect } from 'react'

import isAbsoluteUrl from 'is-absolute-url'
import { iterate } from 'iterare'
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

import styles from './LineDecorator.module.scss'

export interface LineDecoratorProps extends ThemeProps {
    groupedDecorations: Map<{ name: string }, DecorationMapByLine>
    codeViewElements: ReplaySubject<HTMLElement | null>
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number) => HTMLTableCellElement | null
}

/**
 * Component that decorates lines of code and appends line attachments set by extensions
 */
export const LineDecorators = React.memo<LineDecoratorProps>(
    ({ groupedDecorations, getCodeElementFromLineNumber, isLightTheme, codeViewElements }) => {
        const [portalNodes, setPortalNodes] = React.useState<
            Map<HTMLTableCellElement, TextDocumentDecoration[] | undefined>
        >()

        // `LineDecorator` uses `useLayoutEffect` instead of `useEffect` in order to synchronously re-render
        // after mount/decoration updates, but before the browser has painted DOM updates.
        // This prevents users from seeing inconsistent states where changes handled by React have been
        // painted, but DOM manipulation handled by these effects are painted on the next tick.

        // Create portal node and attach to code cell
        useLayoutEffect(() => {
            const addedCells = new Map<HTMLTableCellElement, TextDocumentDecoration[] | undefined>()

            const removeAddedCells = (): void => {
                for (const [cell] of addedCells) {
                    cell.remove()
                }
            }

            // TODO(tj): confirm that this fixes theme toggle decorations bug (probably should, since we have references observable now)
            const subscription = codeViewElements.subscribe(codeView => {
                if (codeView) {
                    const table = codeView.firstElementChild as HTMLTableElement
                    const rows = table.tBodies[0].rows

                    // add extension labels
                    let k = 0
                    for (const [{ name }] of groupedDecorations) {
                        const head = table.tHead || table.createTHead()
                        const hRow = head.rows[0] || head.insertRow()
                        const hCell = hRow.cells[k] || hRow.insertCell(0)
                        hCell.textContent = name
                        addedCells.set(hCell, undefined)
                        k++
                    }

                    console.log(groupedDecorations)

                    // iterate table rows
                    for (let index = 0; index < rows.length; index++) {
                        // add cell for each extension to each row
                        for (const [{ name }, extensionDecorations] of groupedDecorations) {
                            const row = rows[index]

                            // find the existing cell or create a new one if it not exists
                            const cell = row.querySelector<HTMLTableCellElement>(`td.${name}`) || row.insertCell(0)
                            // if (!cell) {
                            // cell = document.createElement('td')
                            cell.classList.add(name)
                            cell.dataset.line = `${index + 1}`
                            cell.dataset.testid = 'line-decoration'
                            cell.dataset.lineDecorationAttachmentPortal = 'true'
                            cell.style.borderRight = '1px solid gray'
                            // row.prepend(cell)
                            // }

                            // get decorations for the 1-based line number
                            const decorations = extensionDecorations.get(index + 1)

                            // add decoration styles to the cell
                            for (const decoration of decorations || []) {
                                const style = decorationStyleForTheme(decoration, isLightTheme)

                                for (const styleProperty of [
                                    'backgroundColor',
                                    'border',
                                    'borderColor',
                                    'borderWidth',
                                ]) {
                                    cell.style[styleProperty] = style[styleProperty]
                                }
                            }

                            // store created cells
                            addedCells.set(cell, decorations)
                        }
                    }
<<<<<<< Updated upstream

=======
>>>>>>> Stashed changes
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
        }, [getCodeElementFromLineNumber, codeViewElements, isLightTheme, groupedDecorations])

        if (!portalNodes?.size) {
            return null
        }

        return iterate(portalNodes)
            .map(([portalRoot, decorations]) =>
                ReactDOM.createPortal(
                    decorations?.filter(property('after', isDefined)).map(decoration => {
                        const attachment = decoration.after
                        const style = decorationAttachmentStyleForTheme(attachment, isLightTheme)

                        return (
                            <LinkOrSpan
                                // Key by content, use index to remove possibility of duplicate keys
                                key={`${decoration.after.contentText ?? decoration.after.hoverMessage ?? ''}-${
                                    portalRoot.dataset.line ?? ''
                                }`}
                                className={styles.lineDecorationAttachment}
                                data-line-decoration-attachment={true}
                                to={attachment.linkURL}
                                data-tooltip={attachment.hoverMessage}
                                // Use target to open external URLs
                                target={attachment.linkURL && isAbsoluteUrl(attachment.linkURL) ? '_blank' : undefined}
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
                                >
                                    {attachment.contentText || ''}
                                </span>
                            </LinkOrSpan>
                        )
                    }),
                    portalRoot
                )
            )
            .toArray()
    }
)

LineDecorators.displayName = 'LineDecorators'
