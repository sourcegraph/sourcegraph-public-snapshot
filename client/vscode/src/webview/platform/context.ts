import { createContext, useContext } from 'react'

import * as Comlink from 'comlink'
import { print } from 'graphql'
import { from, Observable } from 'rxjs'

import { checkOk, GraphQLResult } from '@sourcegraph/http-client'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ExecutableExtension } from '@sourcegraph/shared/src/api/extension/activation'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ExtensionManifest } from '@sourcegraph/shared/src/schema/extensionSchema'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import extensions from '../../../code-intel-extensions.json' // list of extensionID generated by build-inline-extensions script
import { ExtensionCoreAPI } from '../../contract'

import { EventLogger } from './EventLogger'
import { VsceTelemetryService } from './telemetryService'

export interface VSCodePlatformContext
    extends Pick<
        PlatformContext,
        | 'updateSettings'
        | 'settings'
        | 'getGraphQLClient'
        | 'showMessage'
        | 'showInputBox'
        | 'getScriptURLForExtension'
        | 'getStaticExtensions'
        | 'telemetryService'
        | 'clientApplication'
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
        clientApplication: 'other', // TODO add 'vscode-extension' to `clientApplication`,
        getScriptURLForExtension: () => undefined,
        // TODO showInputBox
        // TODO showMessage
        getStaticExtensions: () => getInlineExtensions(),
    }

    return context
}

export interface WebviewPageProps {
    extensionCoreAPI: Comlink.Remote<ExtensionCoreAPI>
    platformContext: VSCodePlatformContext
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

function getInlineExtensions(): Observable<ExecutableExtension[]> {
    const promises: Promise<ExecutableExtension>[] = []

    for (const extensionID of extensions) {
        const { manifestURL, scriptURL } = getURLsForInlineExtension(extensionID)
        promises.push(
            fetch(manifestURL)
                .then(response => checkOk(response).json())
                .then(
                    (manifest: ExtensionManifest): ExecutableExtension => ({
                        id: extensionID,
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
function getURLsForInlineExtension(extensionID: string): { manifestURL: string; scriptURL: string } {
    const extensionsDistributionPath = document.documentElement.dataset.extensionsDistPath!
    const kebabCaseExtensionID = extensionID.replace(/^sourcegraph\//, 'sourcegraph-')
    return {
        manifestURL: `${extensionsDistributionPath}/${kebabCaseExtensionID}/package.json`,
        scriptURL: `${extensionsDistributionPath}/${kebabCaseExtensionID}/extension.js`,
    }
}
