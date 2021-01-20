import React, { useLayoutEffect } from 'react'
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

export interface LineDecoratorProps extends ThemeProps {
    /** 1-based line number */
    line: number
    portalID: string
    decorations: TextDocumentDecoration[]
    codeViewReference: React.MutableRefObject<HTMLElement | null | undefined>
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number) => HTMLTableCellElement | null
}

/**
 * Component that decorates lines of code and appends line attachments set by extensions
 */
export const LineDecorator = React.memo<LineDecoratorProps>(
    ({ getCodeElementFromLineNumber, line, decorations, codeViewReference, portalID, isLightTheme }) => {
        const [portalNode, setPortalNode] = React.useState<HTMLDivElement | null>(null)

        // `LineDecorator` uses `useLayoutEffect` instead of `useEffect` in order to synchronously re-render
        // after mount/decoration updates, but before the browser has painted DOM updates.
        // This prevents users from seeing inconsistent states where changes handled by React have been
        // painted, but DOM manipulation handled by these effects are painted on the next tick.

        // Create portal node and attach to code cell
        useLayoutEffect(() => {
            // code view ref should always be set at this point
            if (codeViewReference.current) {
                const codeCell = getCodeElementFromLineNumber(codeViewReference.current, line)

                const innerPortalNode =
                    portalNode ??
                    // First render, create portalNode
                    (() => {
                        const innerPortalNode = document.createElement('div')
                        innerPortalNode.id = portalID
                        innerPortalNode.classList.add('line-decoration-attachment-portal')
                        return innerPortalNode
                    })()

                if (innerPortalNode.parentElement !== codeCell) {
                    codeCell?.append(innerPortalNode)
                    setPortalNode(innerPortalNode)
                }
            }

            return () => {
                // No portal node to remove on first render
                portalNode?.remove()
            }
        }, [portalNode, getCodeElementFromLineNumber, line, codeViewReference, portalID])

        // Render line decorations
        useLayoutEffect(() => {
            let decoratedElements: HTMLElement[] = []

            // Code view ref should always be set at this point
            if (codeViewReference.current) {
                const codeCell = getCodeElementFromLineNumber(codeViewReference.current, line)
                const row = codeCell?.parentElement as HTMLTableRowElement | undefined

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
                }
            } else {
                decoratedElements = []
            }

            return () => {
                // Clear previous decorations
                for (const element of decoratedElements) {
                    element.style.backgroundColor = ''
                    element.style.border = ''
                    element.style.borderColor = ''
                    element.style.borderWidth = ''
                }
            }
        }, [decorations, codeViewReference, getCodeElementFromLineNumber, isLightTheme, line])

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
