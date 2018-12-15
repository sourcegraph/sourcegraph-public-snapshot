import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import * as React from 'react'
import { render } from 'react-dom'
import { ContributableMenu } from '../../../../../shared/src/api/protocol'
import { CommandListPopoverButton } from '../../../../../shared/src/commandPalette/CommandList'
import { Notifications } from '../../../../../shared/src/notifications/Notifications'

import { DOMFunctions } from '@sourcegraph/codeintellify'
import * as H from 'history'
import {
    decorationAttachmentStyleForTheme,
    decorationStyleForTheme,
} from '../../../../../shared/src/api/client/services/decoration'
import {
    createController as createExtensionsController,
    ExtensionsControllerProps,
} from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { createPlatformContext } from '../../platform/context'
import { GlobalDebug } from '../../shared/components/GlobalDebug'
import { ShortcutProvider } from '../../shared/components/ShortcutProvider'
import { getGlobalDebugMount } from '../github/extensions'
import { CodeHost } from './code_intelligence'

/**
 * Initializes extensions for a page. It creates the {@link PlatformContext} and extensions controller.
 *
 * If the "Use extensions" feature flag is enabled (or always for Sourcegraph.com), it injects the command palette.
 * If extensions are not supported by the associated Sourcegraph instance, the extensions controller will behave as
 * though no individual extensions are enabled, which makes it effectively a noop.
 */
export function initializeExtensions({
    getCommandPaletteMount,
}: Pick<CodeHost, 'getCommandPaletteMount'>): PlatformContextProps & ExtensionsControllerProps {
    const platformContext = createPlatformContext()
    const extensionsController = createExtensionsController(platformContext)
    const history = H.createBrowserHistory()

    if (getCommandPaletteMount) {
        render(
            <ShortcutProvider>
                <CommandListPopoverButton
                    extensionsController={extensionsController}
                    menu={ContributableMenu.CommandPalette}
                    platformContext={platformContext}
                    autoFocus={false}
                    location={history.location}
                />
                <Notifications extensionsController={extensionsController} />
            </ShortcutProvider>,
            getCommandPaletteMount()
        )
    }

    render(
        <GlobalDebug
            extensionsController={extensionsController}
            location={history.location}
            platformContext={platformContext}
        />,
        getGlobalDebugMount()
    )

    return { platformContext, extensionsController }
}

const IS_LIGHT_THEME = true // assume all code hosts have a light theme (correct for now)

const groupByLine = (decorations: TextDocumentDecoration[]) => {
    const grouped = new Map<number, TextDocumentDecoration[]>()
    for (const d of decorations) {
        const lineNumber = d.range.start.line + 1
        const decorationsForLine = grouped.get(lineNumber)
        if (!decorationsForLine) {
            grouped.set(lineNumber, [d])
        } else {
            decorationsForLine.push(d)
        }
    }
    return grouped
}

export const cleanupDecorations = (dom: DOMFunctions, codeView: HTMLElement, lines: number[]): void => {
    for (const lineNumber of lines) {
        const codeElement = dom.getCodeElementFromLineNumber(codeView, lineNumber)
        if (!codeElement) {
            continue
        }
        codeElement.style.backgroundColor = null
        const previousDecorations = codeElement.querySelectorAll('.line-decoration-attachment')
        for (const d of previousDecorations) {
            d.remove()
        }
    }
}

/**
 * Applies a decoration to a code view. This doesn't work with diff views yet.
 */
export const applyDecorations = (
    dom: DOMFunctions,
    codeView: HTMLElement,
    decorations: TextDocumentDecoration[],
    previousDecorations: number[]
): number[] => {
    cleanupDecorations(dom, codeView, previousDecorations)
    const decorationsByLine = groupByLine(decorations)
    for (const [lineNumber, decorationsForLine] of decorationsByLine) {
        const codeElement = dom.getCodeElementFromLineNumber(codeView, lineNumber)
        if (!codeElement) {
            throw new Error(`Unable to find code element for line ${lineNumber}`)
        }
        for (const decoration of decorationsForLine) {
            const style = decorationStyleForTheme(decoration, IS_LIGHT_THEME)
            if (style.backgroundColor) {
                codeElement.style.backgroundColor = style.backgroundColor
            }

            if (decoration.after) {
                const style = decorationAttachmentStyleForTheme(decoration.after, IS_LIGHT_THEME)

                const linkTo = (url: string) => (e: HTMLElement): HTMLElement => {
                    const link = document.createElement('a')
                    link.setAttribute('href', url)

                    // External URLs should open in a new tab, whereas relative URLs
                    // should not.
                    link.setAttribute('target', /^https?:\/\//.test(url) ? '_blank' : '')

                    // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
                    link.setAttribute('rel', 'noreferrer noopener')

                    link.style.color = style.color || null
                    link.appendChild(e)
                    return link
                }

                const after = document.createElement('span')
                after.style.backgroundColor = style.backgroundColor || null
                after.textContent = decoration.after.contentText || null
                after.title = decoration.after.hoverMessage || ''

                const annotation = decoration.after.linkURL ? linkTo(decoration.after.linkURL)(after) : after
                annotation.className = 'sourcegraph-extension-element line-decoration-attachment'
                codeElement.appendChild(annotation)
            }
        }
    }
    return [...decorationsByLine.keys()]
}
