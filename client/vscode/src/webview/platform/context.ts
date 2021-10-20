import * as Comlink from 'comlink'
import { from } from 'rxjs'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { SourcegraphVSCodeExtensionAPI } from '../contract'

export interface VSCodePlatformContext extends Pick<PlatformContext, 'requestGraphQL'> {}

export function createPlatformContext(
    sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI>
): VSCodePlatformContext {
    const context: VSCodePlatformContext = {
        requestGraphQL({ request, variables }) {
            return from(sourcegraphVSCodeExtensionAPI.requestGraphQL(request, variables))
        },
    }

    return context
}
