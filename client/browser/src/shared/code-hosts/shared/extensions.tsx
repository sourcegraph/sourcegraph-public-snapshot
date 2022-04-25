import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import { render } from 'react-dom'

import { ContributableMenu } from '@sourcegraph/client-api'
import { DiffPart } from '@sourcegraph/codeintellify'
import { isExternalLink } from '@sourcegraph/common'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import {
    decorationAttachmentStyleForTheme,
    DecorationMapByLine,
    decorationStyleForTheme,
    groupDecorationsByLine,
} from '@sourcegraph/shared/src/api/extension/api/decorations'
import {
    CommandListPopoverButton,
    CommandListPopoverButtonProps,
} from '@sourcegraph/shared/src/commandPalette/CommandList'
import {
    createController as createExtensionsController,
    ExtensionsControllerProps,
} from '@sourcegraph/shared/src/extensions/controller'
import { UnbrandedNotificationItemStyleProps } from '@sourcegraph/shared/src/notifications/NotificationItem'
import { Notifications } from '@sourcegraph/shared/src/notifications/Notifications'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { GlobalDebug } from '../../components/GlobalDebug'
import { ShortcutProvider } from '../../components/ShortcutProvider'
import { createPlatformContext, SourcegraphIntegrationURLs, BrowserPlatformContext } from '../../platform/context'

import { CodeHost } from './codeHost'
import { DOMFunctions } from './codeViews'

/**
 * Initializes extensions for a page. It creates the {@link PlatformContext} and extensions controller.
 *
 */
export function initializeExtensions(
    { urlToFile }: Pick<CodeHost, 'urlToFile'>,
    urls: SourcegraphIntegrationURLs,
    isExtension: boolean
): { platformContext: BrowserPlatformContext } & ExtensionsControllerProps {
    const platformContext = createPlatformContext({ urlToFile }, urls, isExtension)
    const extensionsController = createExtensionsController(platformContext)
    return { platformContext, extensionsController }
}

interface InjectProps
    extends PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sideloadedExtensionURL' | 'sourcegraphURL'>,
        ExtensionsControllerProps {
    history: H.History
    render: typeof render
}

interface RenderCommandPaletteProps
    extends TelemetryProps,
        InjectProps,
        Pick<CommandListPopoverButtonProps, 'inputClassName' | 'popoverClassName' | 'popoverInnerClassName'> {
    notificationClassNames: UnbrandedNotificationItemStyleProps['notificationItemClassNames']
}

export const renderCommandPalette = ({
    extensionsController,
    history,
    render,
    ...props
}: RenderCommandPaletteProps) => (mount: HTMLElement): void => {
    render(
        <ShortcutProvider>
            <CommandListPopoverButton
                {...props}
                popoverClassName={classNames('command-list-popover', props.popoverClassName)}
                popoverInnerClassName={props.popoverInnerClassName}
                menu={ContributableMenu.CommandPalette}
                extensionsController={extensionsController}
                location={history.location}
            />
            <Notifications
                extensionsController={extensionsController}
                notificationItemStyleProps={{
                    notificationItemClassNames: props.notificationClassNames,
                }}
            />
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

const cleanupDecorationsForCodeElement = (codeElement: HTMLElement, part: DiffPart | undefined): void => {
    codeElement.style.backgroundColor = ''
    const previousAttachments = codeElement.querySelectorAll(
        `[data-line-decoration-attachment][data-part=${String(part)}]`
    )
    for (const attachment of previousAttachments) {
        attachment.remove()
    }
}

const cleanupDecorationsForLineElement = (lineElement: HTMLElement): void => {
    lineElement.style.backgroundColor = ''
}

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
    isLightTheme: boolean,
    previousIsLightTheme: boolean,
    part?: DiffPart
): DecorationMapByLine => {
    const decorationsByLine = groupDecorationsByLine(decorations)
    // Clean up lines that now don't have decorations anymore
    for (const lineNumber of previousDecorations.keys()) {
        if (!decorationsByLine.has(lineNumber)) {
            const codeElement = dom.getCodeElementFromLineNumber(codeView, lineNumber, part)
            if (codeElement) {
                cleanupDecorationsForCodeElement(codeElement, part)
            }
            const lineElement = dom.getLineElementFromLineNumber(codeView, lineNumber, part)
            if (lineElement) {
                cleanupDecorationsForLineElement(lineElement)
            }
        }
    }
    for (const [lineNumber, decorationsForLine] of decorationsByLine) {
        const previousDecorationsForLine = previousDecorations.get(lineNumber)
        if (previousIsLightTheme === isLightTheme && isEqual(decorationsForLine, previousDecorationsForLine)) {
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

        // Clean up previous decorations
        // Sometimes these can be there even if we cleaned them up if
        // the code host snapshotted the DOM before removal of the code view
        // (happens on GitHub when switching tabs on a PR)
        cleanupDecorationsForCodeElement(codeElement, part)
        cleanupDecorationsForLineElement(lineElement)

        for (const decoration of decorationsForLine) {
            const style = decorationStyleForTheme(decoration, isLightTheme)
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
                const style = decorationAttachmentStyleForTheme(decoration.after, isLightTheme)

                const linkTo = (url: string) => (element: HTMLElement): HTMLElement => {
                    const link = document.createElement('a')
                    link.setAttribute('href', url)

                    // External URLs should open in a new tab, whereas relative URLs
                    // should not.
                    link.setAttribute('target', isExternalLink(url) ? '_blank' : '')

                    // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
                    link.setAttribute('rel', 'noreferrer noopener')

                    link.style.color = style.color || ''
                    link.append(element)
                    return link
                }

                const after = document.createElement('span')
                after.style.color = style.color || ''
                after.style.backgroundColor = style.backgroundColor || ''
                after.textContent = decoration.after.contentText || null
                if (decoration.after.hoverMessage) {
                    after.title = decoration.after.hoverMessage
                }

                const annotation = decoration.after.linkURL ? linkTo(decoration.after.linkURL)(after) : after
                annotation.dataset.part = String(part)
                annotation.className = 'sourcegraph-extension-element'
                annotation.dataset.lineDecorationAttachment = 'true'
                codeElement.append(annotation)
            }
        }
    }
    return decorationsByLine
}
