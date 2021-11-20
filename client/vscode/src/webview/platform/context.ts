import * as Comlink from 'comlink'
import { from } from 'rxjs'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { SourcegraphVSCodeExtensionAPI } from '../contract'

import { vscodeTelemetryService } from './telemetryService'

export interface VSCodePlatformContext extends Pick<PlatformContext, 'requestGraphQL' | 'settings'> {
    telemetryService: TelemetryService
}

export function createPlatformContext(
    sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI>
): VSCodePlatformContext {
    const context: VSCodePlatformContext = {
        requestGraphQL({ request, variables }) {
            return from(sourcegraphVSCodeExtensionAPI.requestGraphQL(request, variables))
        },
        // TODO: refresh settings in extension every hour that a search panel is created.
        settings: wrapRemoteObservable(sourcegraphVSCodeExtensionAPI.getSettings()),
        telemetryService: vscodeTelemetryService,
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
