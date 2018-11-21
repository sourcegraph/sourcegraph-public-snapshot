import * as React from 'react'
import { render } from 'react-dom'
import { combineLatest, from, Observable, Unsubscribable } from 'rxjs'
import { filter, take } from 'rxjs/operators'
import { ContributableMenu } from '../../../../../shared/src/api/protocol'
import { TextDocumentDecoration } from '../../../../../shared/src/api/protocol/plainTypes'
import { CommandListPopoverButton } from '../../../../../shared/src/commandPalette/CommandList'
import { Controller as ClientController, createController } from '../../../../../shared/src/extensions/controller'
import { Notifications } from '../../../../../shared/src/notifications/Notifications'

import { DOMFunctions } from '@sourcegraph/codeintellify'
import * as H from 'history'
import { Environment } from '../../../../../shared/src/api/client/environment'
import {
    decorationAttachmentStyleForTheme,
    decorationStyleForTheme,
} from '../../../../../shared/src/api/client/providers/decoration'
import { viewerConfiguredExtensions } from '../../../../../shared/src/extensions/helpers'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { isSettingsValid } from '../../../../../shared/src/settings/settings'
import { createPlatformContext } from '../../shared/backend/extensions'
import { GlobalDebug } from '../../shared/components/GlobalDebug'
import { ShortcutProvider } from '../../shared/components/ShortcutProvider'
import { sourcegraphUrl } from '../../shared/util/context'
import { getGlobalDebugMount } from '../github/extensions'
import { MountGetter } from './code_intelligence'

export interface Controllers {
    platformContext: PlatformContext
    extensionsController: ClientController
}

function createControllers(environment: Observable<Pick<Environment, 'roots' | 'visibleTextDocuments'>>): Controllers {
    const platformContext = createPlatformContext(sourcegraphUrl)
    const extensionsController = createController(platformContext)

    combineLatest(
        viewerConfiguredExtensions(platformContext),
        from(platformContext.settingsCascade).pipe(filter(isSettingsValid)),
        environment
    ).subscribe(([extensions, configuration, { roots, visibleTextDocuments }]) => {
        from(extensionsController.environment)
            .pipe(take(1))
            .subscribe(({ context }) => {
                extensionsController.setEnvironment({
                    roots,
                    extensions,
                    configuration,
                    visibleTextDocuments,
                    context,
                })
            })
    })

    return { platformContext, extensionsController }
}

/**
 * Initializes extensions for a page. It creates the controllers and injects the command palette.
 */
export function initializeExtensions(
    getCommandPaletteMount: MountGetter,
    environment: Observable<Pick<Environment, 'roots' | 'visibleTextDocuments'>>
): Controllers {
    const { platformContext, extensionsController } = createControllers(environment)
    const history = H.createBrowserHistory()

    render(
        <ShortcutProvider>
            <CommandListPopoverButton
                extensionsController={extensionsController}
                menu={ContributableMenu.CommandPalette}
                platformContext={platformContext}
                location={history.location}
            />
            <Notifications extensionsController={extensionsController} />
        </ShortcutProvider>,
        getCommandPaletteMount()
    )

    render(
        <GlobalDebug extensionsController={extensionsController} location={history.location} />,
        getGlobalDebugMount()
    )

    return { platformContext, extensionsController }
}

const combineUnsubscribables = (...unsubscribables: Unsubscribable[]): Unsubscribable => ({
    unsubscribe: () => {
        for (const unsubscribable of unsubscribables) {
            unsubscribable.unsubscribe()
        }
    },
})

const IS_LIGHT_THEME = true // assume all code hosts have a light theme (correct for now)

/**
 * Applies a decoration to a code view. This doesn't work with diff views yet.
 */
export const applyDecoration = (
    dom: DOMFunctions,
    {
        codeView,
        decoration,
    }: {
        codeView: HTMLElement
        decoration: TextDocumentDecoration
    }
): Unsubscribable => {
    const unsubscribables: Unsubscribable[] = []

    const lineNumber = decoration.range.start.line + 1
    const codeElement = dom.getCodeElementFromLineNumber(codeView, lineNumber)
    if (!codeElement) {
        throw new Error(`Unable to find code element for line ${lineNumber}`)
    }

    const style = decorationStyleForTheme(decoration, IS_LIGHT_THEME)
    if (style.backgroundColor) {
        codeElement.style.backgroundColor = style.backgroundColor
        unsubscribables.push({
            unsubscribe: () => {
                codeElement.style.backgroundColor = null
            },
        })
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

        unsubscribables.push({
            unsubscribe: () => {
                annotation.remove()
            },
        })
    }
    return combineUnsubscribables(...unsubscribables)
}
