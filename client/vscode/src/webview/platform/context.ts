import { createContext, useContext } from 'react'

import * as Comlink from 'comlink'
import { print } from 'graphql'
import { BehaviorSubject, from, Observable, Subscribable } from 'rxjs'
import { checkOk, GraphQLResult } from '@sourcegraph/http-client'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { DeprecatedTooltipController } from '@sourcegraph/wildcard'

import { extensions } from '../../../bundled-code-intel-extensions.json'

import { ExtensionCoreAPI } from '../../contract'

import { EventLogger } from './EventLogger'
import { VsceTelemetryService } from './telemetryService'
import { ExecutableExtension } from '@sourcegraph/shared/src/api/extension/activation'
import { ExtensionManifest } from '@sourcegraph/shared/src/schema/extensionSchema'

export interface VSCodePlatformContext
    extends Pick<
        PlatformContext,
        | 'updateSettings'
        | 'settings'
        | 'getGraphQLClient'
        | 'showMessage'
        | 'showInputBox'
        | 'sideloadedExtensionURL'
        | 'getScriptURLForExtension'
        | 'getStaticExtensions'
        | 'telemetryService'
        | 'clientApplication'
        | 'forceUpdateTooltip'
    > {
    // Ensure telemetryService is non-nullable.
    telemetryService: VsceTelemetryService
    requestGraphQL: <R, V = object>(options: {
        request: string
        variables: V
        mightContainPrivateInfo: boolean
        overrideAccessToken?: string
        overrideSourcegraphURL?: string
    }) => Observable<GraphQLResult<R>>
}

export function createPlatformContext(extensionCoreAPI: Comlink.Remote<ExtensionCoreAPI>): VSCodePlatformContext {
    const context: VSCodePlatformContext = {
        requestGraphQL({ request, variables, overrideAccessToken, overrideSourcegraphURL }) {
            return from(
                extensionCoreAPI.requestGraphQL(request, variables, overrideAccessToken, overrideSourcegraphURL)
            )
        },
        // TODO add true Apollo Client support for v2
        getGraphQLClient: () =>
            Promise.resolve({
                watchQuery: ({ variables, query }) =>
                    from(extensionCoreAPI.requestGraphQL(print(query), variables)) as any,
            }),
        settings: wrapRemoteObservable(extensionCoreAPI.observeSourcegraphSettings()),
        // TODO: implement GQL mutation, settings refresh (called by extensions, impl w/ ext. host).
        updateSettings: () => Promise.resolve(),
        telemetryService: new EventLogger(extensionCoreAPI),
        sideloadedExtensionURL: new BehaviorSubject<string | null>(null),
        clientApplication: 'other', // TODO add 'vscode-extension' to `clientApplication`,
        getScriptURLForExtension: () => undefined,
        forceUpdateTooltip: () => DeprecatedTooltipController.forceUpdate(),
        // TODO showInputBox
        // TODO showMessage
        getStaticExtensions: () => {
            return getInlineExtensions()
        },
    }

    return context
}

export interface WebviewPageProps {
    extensionCoreAPI: Comlink.Remote<ExtensionCoreAPI>
    platformContext: VSCodePlatformContext
    theme: 'theme-dark' | 'theme-light'
    authenticatedUser: AuthenticatedUser | null
    settingsCascade: SettingsCascadeOrError
    instanceURL: string
}

// Webview page context. Used to pass to aliased components.
export const WebviewPageContext = createContext<WebviewPageProps | undefined>(undefined)

export function useWebviewPageContext(): WebviewPageProps {
    const context = useContext(WebviewPageContext)

    if (context === undefined) {
        throw new Error('useWebviewPageContext must be used within a WebviewPageContextProvider')
    }

    return context
}

function getInlineExtensions(): Subscribable<ExecutableExtension[]> {
    const promises: Promise<ExecutableExtension>[] = []

    for (let extensionName of extensions) {
        const { manifestURL, scriptURL } = getURLsForInlineExtension(extensionName)
        promises.push(
            fetch(manifestURL)
                .then(response => checkOk(response).json())
                .then(
                    (manifest: ExtensionManifest): ExecutableExtension => ({
                        id: `sourcegraph/${extensionName}`,
                        manifest,
                        scriptURL,
                    })
                )
        )
    }

    return from(Promise.all(promises))
}

/**
 * Get the manifest URL and script URL for a Sourcegraph extension which is inline (bundled with the browser add-on).
 */
function getURLsForInlineExtension(extensionName: string): { manifestURL: string; scriptURL: string } {
    const extensionsDistPath = document.documentElement.dataset.extensionsDistPath!
    return {
        manifestURL: `${extensionsDistPath}/${extensionName}/package.json`,
        scriptURL: `${extensionsDistPath}/${extensionName}/extension.js`,
    }
}
