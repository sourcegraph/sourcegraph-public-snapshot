import React, { useLayoutEffect, useMemo, useState } from 'react'
import ReactDOM from 'react-dom'
import isAbsoluteUrl from 'is-absolute-url'
import {
    decorationAttachmentStyleForTheme,
    decorationStyleForTheme,
} from '../../../../shared/src/api/client/services/decoration'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { ThemeProps } from '../../../../shared/src/theme'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { isDefined, property } from '../../../../shared/src/util/types'
import { Observable, ReplaySubject } from 'rxjs'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { filter, finalize, map, tap } from 'rxjs/operators'

export interface LineDecoratorProps extends ThemeProps {
    /** 1-based line number */
    line: number
    portalID: string
    decorations: TextDocumentDecoration[]
    codeViewElements: ReplaySubject<HTMLElement | null>
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number) => HTMLTableCellElement | null
}

/**
 * Component that decorates lines of code and appends line attachments set by extensions
 */
export const LineDecorator = React.memo<LineDecoratorProps>(
    ({ getCodeElementFromLineNumber, line, decorations, portalID, isLightTheme, codeViewElements }) => {
        const [portalNode, setPortalNode] = React.useState<HTMLDivElement | null>(null)

        // `LineDecorator` uses `useLayoutEffect` instead of `useEffect` in order to synchronously re-render
        // after mount/decoration updates, but before the browser has painted DOM updates.
        // This prevents users from seeing inconsistent states where changes handled by React have been
        // painted, but DOM manipulation handled by these effects are painted on the next tick.

        // Create portal node and attach to code cell
        useLayoutEffect(() => {
            let innerPortalNode: HTMLDivElement | null = null
            let decoratedElements: HTMLElement[] = []

            // TODO(tj): confirm that this fixes theme toggle decorations bug (probably should, since we have references observable now)
            const subscription = codeViewElements.subscribe(codeView => {
                if (codeView) {
                    const codeCell = getCodeElementFromLineNumber(codeView, line)
                    const row = codeCell?.parentElement as HTMLTableRowElement | undefined

                    // Clear previous decoration styles if exists
                    for (const element of decoratedElements) {
                        element.style.backgroundColor = ''
                        element.style.border = ''
                        element.style.borderColor = ''
                        element.style.borderWidth = ''
                    }

                    // Apply line decoration styles
                    if (row) {
                        for (const decoration of decorations) {
                            let decorated = false
                            const style = decorationStyleForTheme(decoration, isLightTheme)
                            if (style.backgroundColor) {
                                row.style.backgroundColor = style.backgroundColor
                                decorated = true
                            }
                            if (style.border) {
                                row.style.border = style.border
                                decorated = true
                            }
                            if (style.borderColor) {
                                row.style.borderColor = style.borderColor
                                decorated = true
                            }
                            if (style.borderWidth) {
                                row.style.borderWidth = style.borderWidth
                                decorated = true
                            }
                            if (decorated) {
                                decoratedElements.push(row)
                            }
                        }
                    } else {
                        decoratedElements = []
                    }

                    // Create portal
                    // Remove existing portal node if exists
                    innerPortalNode?.remove()
                    innerPortalNode = document.createElement('div')
                    innerPortalNode.id = portalID
                    innerPortalNode.classList.add('line-decoration-attachment-portal')
                    codeCell?.append(innerPortalNode)
                    setPortalNode(innerPortalNode)
                } else {
                    // code view ref passed `null`, so element is leaving DOM
                    innerPortalNode?.remove()
                    for (const element of decoratedElements) {
                        element.style.backgroundColor = ''
                        element.style.border = ''
                        element.style.borderColor = ''
                        element.style.borderWidth = ''
                    }
                }
            })

            return () => {
                subscription.unsubscribe()
                innerPortalNode?.remove()
                for (const element of decoratedElements) {
                    element.style.backgroundColor = ''
                    element.style.border = ''
                    element.style.borderColor = ''
                    element.style.borderWidth = ''
                }
            }
        }, [getCodeElementFromLineNumber, codeViewElements, line, portalID, decorations, isLightTheme])

        if (!portalNode) {
            return null
        }

        // Render decoration attachments into portal
        return ReactDOM.createPortal(
            decorations?.filter(property('after', isDefined)).map((decoration, index) => {
                const attachment = decoration.after
                const style = decorationAttachmentStyleForTheme(attachment, isLightTheme)

                return (
                    <LinkOrSpan
                        // Key by content, use index to remove possibility of duplicate keys
                        key={`${decoration.after.contentText ?? decoration.after.hoverMessage ?? ''}-${index}`}
                        className="line-decoration-attachment"
                        to={attachment.linkURL}
                        data-tooltip={attachment.hoverMessage}
                        // Use target to open external URLs
                        target={attachment.linkURL && isAbsoluteUrl(attachment.linkURL) ? '_blank' : undefined}
                        // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
                        rel="noreferrer noopener"
                    >
                        <span
                            className="line-decoration-attachment__contents"
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
            portalNode
        )
    }
)
