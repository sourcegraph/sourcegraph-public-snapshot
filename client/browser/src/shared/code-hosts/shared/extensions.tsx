import type { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/createSyncLoadedController'
import type { TelemetryRecorderProvider } from '@sourcegraph/shared/src/telemetry'

import type { GraphQLHelpers } from '../../backend/requestGraphQl'
import {
    createPlatformContext,
    type SourcegraphIntegrationURLs,
    type BrowserPlatformContext,
} from '../../platform/context'

import type { CodeHost } from './codeHost'

/**
 * Initializes extensions for a page. It creates the {@link PlatformContext} and extensions controller.
 *
 */
export function initializeExtensions(
    graphql: GraphQLHelpers,
    { urlToFile }: Pick<CodeHost, 'urlToFile'>,
    urls: SourcegraphIntegrationURLs,
    telemetryRecorderProvider: TelemetryRecorderProvider
): {
    platformContext: BrowserPlatformContext
} & ExtensionsControllerProps {
    const platformContext = createPlatformContext(graphql, { urlToFile }, urls, telemetryRecorderProvider)
    const extensionsController = createExtensionsController(platformContext)
    return { platformContext, extensionsController }
}
