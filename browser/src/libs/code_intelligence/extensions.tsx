import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import * as React from 'react'
import { render } from 'react-dom'
import { ContributableMenu } from '../../../../shared/src/api/protocol'
import {
    CommandListPopoverButton,
    CommandListPopoverButtonProps,
} from '../../../../shared/src/commandPalette/CommandList'
import { Notifications } from '../../../../shared/src/notifications/Notifications'

import { DiffPart } from '@sourcegraph/codeintellify'
import * as H from 'history'
import { isEqual } from 'lodash'
import {
    decorationAttachmentStyleForTheme,
    decorationStyleForTheme,
} from '../../../../shared/src/api/client/services/decoration'
import {
    createController as createExtensionsController,
    ExtensionsControllerProps,
} from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { createPlatformContext } from '../../platform/context'
import { GlobalDebug } from '../../shared/components/GlobalDebug'
import { ShortcutProvider } from '../../shared/components/ShortcutProvider'
import { CodeHost } from './code_intelligence'
import { DOMFunctions } from './code_views'

/**
 * Initializes extensions for a page. It creates the {@link PlatformContext} and extensions controller.
 *
 */
export function initializeExtensions(
    { urlToFile, getContext }: Pick<CodeHost, 'urlToFile' | 'getContext'>,
    sourcegraphURL: string,
    isExtension: boolean
): PlatformContextProps & ExtensionsControllerProps {
    const platformContext = createPlatformContext({ urlToFile, getContext }, sourcegraphURL, isExtension)
    const extensionsController = createExtensionsController(platformContext)
    return { platformContext, extensionsController }
}

interface InjectProps
    extends PlatformContextProps<'forceUpdateTooltip' | 'sideloadedExtensionURL'>,
        ExtensionsControllerProps {
    history: H.History
    render: typeof render
}

export const renderCommandPalette = ({
    extensionsController,
    history,
    render,
    ...props
}: TelemetryProps & InjectProps & Pick<CommandListPopoverButtonProps, 'inputClassName' | 'popoverClassName'>) => (
    mount: HTMLElement
): void => {
    render(
        <ShortcutProvider>
            <CommandListPopoverButton
                {...props}
                menu={ContributableMenu.CommandPalette}
                extensionsController={extensionsController}
                location={history.location}
            />
            <Notifications extensionsController={extensionsController} />
        </ShortcutProvider>,
        mount
    )
}

export const renderGlobalDebug = ({
    extensionsController,
    platformContext,
    history,
    render,
    sourcegraphURL,
}: InjectProps & { sourcegraphURL: string; showGlobalDebug?: boolean }) => (mount: HTMLElement): void => {
    render(
        <GlobalDebug
            extensionsController={extensionsController}
            location={history.location}
            platformContext={platformContext}
            sourcegraphURL={sourcegraphURL}
        />,
        mount
    )
}

const IS_LIGHT_THEME = true // assume all code hosts have a light theme (correct for now)

/**
 * @returns Map from line number to non-empty array of TextDocumentDecoration for that line
 */
const groupByLine = (decorations: TextDocumentDecoration[]): Map<number, TextDocumentDecoration[]> => {
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

const cleanupDecorationsForCodeElement = (codeElement: HTMLElement): void => {
    codeElement.style.backgroundColor = null
    const previousAttachments = codeElement.querySelectorAll('.line-decoration-attachment')
    for (const attachment of previousAttachments) {
        attachment.remove()
    }
}

const cleanupDecorationsForLineElement = (lineElement: HTMLElement): void => {
    lineElement.style.backgroundColor = null
}

export type DecorationMapByLine = Map<number, TextDocumentDecoration[]>

/**
 * Applies a decoration to a code view. This doesn't work with diff views yet.
 *
 * @returns New decorations, grouped by line number
 */
export const applyDecorations = (
    dom: DOMFunctions,
    codeView: HTMLElement,
    decorations: TextDocumentDecoration[],
    previousDecorations: DecorationMapByLine,
    part?: DiffPart
): DecorationMapByLine => {
    const decorationsByLine = groupByLine(decorations)
    // Clean up lines that now don't have decorations anymore
    for (const lineNumber of previousDecorations.keys()) {
        if (!decorationsByLine.has(lineNumber)) {
            const codeElement = dom.getCodeElementFromLineNumber(codeView, lineNumber)
            if (codeElement) {
                cleanupDecorationsForCodeElement(codeElement)
            }
            const lineElement = dom.getLineElementFromLineNumber(codeView, lineNumber)
            if (lineElement) {
                cleanupDecorationsForLineElement(lineElement)
            }
        }
    }
    for (const [lineNumber, decorationsForLine] of decorationsByLine) {
        const previousDecorationsForLine = previousDecorations.get(lineNumber)
        if (isEqual(decorationsForLine, previousDecorationsForLine)) {
            console.log('Skipping decoration in line', lineNumber)
            // No change in this line
            continue
        }
        const codeElement = dom.getCodeElementFromLineNumber(codeView, lineNumber, part)
        if (!codeElement) {
            if (part === undefined) {
                throw new Error(`Unable to find code element for line ${lineNumber}`)
            }
            // In diffs it's normal that many lines are not visible
            continue
        }
        const lineElement = dom.getLineElementFromLineNumber(codeView, lineNumber, part)
        if (!lineElement) {
            if (part === undefined) {
                throw new Error(`Could not find line element for line ${lineNumber}`)
            }
            // In diffs it's normal that many lines are not visible
            continue
        }
        // Clean up previous decorations if this line had some
        if (previousDecorationsForLine) {
            cleanupDecorationsForCodeElement(codeElement)
            cleanupDecorationsForLineElement(lineElement)
        }
        for (const decoration of decorationsForLine) {
            const style = decorationStyleForTheme(decoration, IS_LIGHT_THEME)
            if (style.backgroundColor) {
                let backgroundElement: HTMLElement
                if (decoration.isWholeLine) {
                    backgroundElement = lineElement
                } else {
                    backgroundElement = codeElement
                }
                backgroundElement.style.backgroundColor = style.backgroundColor
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
    return decorationsByLine
}
