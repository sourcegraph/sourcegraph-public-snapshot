import type { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/createSyncLoadedController'

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
    { urlToFile }: Pick<CodeHost, 'urlToFile'>,
    urls: SourcegraphIntegrationURLs,
    isExtension: boolean
): { platformContext: BrowserPlatformContext } & ExtensionsControllerProps {
    const platformContext = createPlatformContext({ urlToFile }, urls, isExtension)
    const extensionsController = createExtensionsController(platformContext)
    return { platformContext, extensionsController }
}
