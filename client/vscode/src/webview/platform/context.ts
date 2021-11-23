import * as Comlink from 'comlink'
import { print } from 'graphql'
import { BehaviorSubject, from } from 'rxjs'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { SourcegraphVSCodeExtensionAPI } from '../contract'

import { vscodeTelemetryService } from './telemetryService'

export interface VSCodePlatformContext
    extends Pick<
        PlatformContext,
        | 'updateSettings'
        | 'settings'
        | 'getGraphQLClient'
        | 'requestGraphQL'
        | 'showMessage'
        | 'showInputBox'
        | 'sideloadedExtensionURL'
        | 'getScriptURLForExtension'
        | 'getStaticExtensions'
        | 'telemetryService'
        | 'clientApplication'
    > {
    // Ensure telemetryService is non-nullable.
    telemetryService: TelemetryService
}

export function createPlatformContext(
    sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI>
): VSCodePlatformContext {
    const context: VSCodePlatformContext = {
        requestGraphQL({ request, variables }) {
            return from(sourcegraphVSCodeExtensionAPI.requestGraphQL(request, variables))
        },
        getGraphQLClient: () =>
            Promise.resolve({
                watchQuery: ({ variables, query }) =>
                    from(sourcegraphVSCodeExtensionAPI.requestGraphQL(print(query), variables)) as any,
            }),
        // TODO: refresh settings in extension every hour that a search panel is created/file is opened?
        // button to refresh settings?
        settings: wrapRemoteObservable(sourcegraphVSCodeExtensionAPI.getSettings()),
        // TODO: implement GQL mutation, settings refresh
        updateSettings: () => Promise.resolve(),
        telemetryService: vscodeTelemetryService,
        sideloadedExtensionURL: new BehaviorSubject<string | null>(null),
        clientApplication: 'other', // TODO add 'vscode-extension' to `clientApplication`,
        getScriptURLForExtension: () => undefined,
        // TODO showMessage
        // TODO showInputBox
    }

    // Any state that needs to be shared between webview instances (search panels, search sidebar)
    // should live in the extension context, read through `SourcegraphVSCodeExtensionAPI`.

    return context
}

export interface WebviewPageProps {
    sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI>
    platformContext: VSCodePlatformContext
    theme: 'theme-dark' | 'theme-light'
}
