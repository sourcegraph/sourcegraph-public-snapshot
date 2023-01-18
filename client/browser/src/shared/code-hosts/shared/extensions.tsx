import classNames from 'classnames'
import * as H from 'history'
import { Renderer } from 'react-dom'

import { ContributableMenu } from '@sourcegraph/client-api'
import {
    CommandListPopoverButton,
    CommandListPopoverButtonProps,
} from '@sourcegraph/shared/src/commandPalette/CommandList'
import {
    ExtensionsControllerProps,
    RequiredExtensionsControllerProps,
} from '@sourcegraph/shared/src/extensions/controller'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/createSyncLoadedController'
import { UnbrandedNotificationItemStyleProps } from '@sourcegraph/shared/src/notifications/NotificationItem'
import { Notifications } from '@sourcegraph/shared/src/notifications/Notifications'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ShortcutProvider } from '../../components/ShortcutProvider'
import { createPlatformContext, SourcegraphIntegrationURLs, BrowserPlatformContext } from '../../platform/context'

import { CodeHost } from './codeHost'

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

interface InjectProps extends PlatformContextProps<'settings' | 'sourcegraphURL'>, RequiredExtensionsControllerProps {
    history: H.History
    render: Renderer
}

interface RenderCommandPaletteProps
    extends TelemetryProps,
        InjectProps,
        Pick<CommandListPopoverButtonProps, 'inputClassName' | 'popoverClassName'> {
    notificationClassNames: UnbrandedNotificationItemStyleProps['notificationItemClassNames']
}

export const renderCommandPalette =
    ({ extensionsController, history, render, ...props }: RenderCommandPaletteProps) =>
    (mount: HTMLElement): void => {
        render(
            <ShortcutProvider>
                <CommandListPopoverButton
                    {...props}
                    popoverClassName={classNames('command-list-popover', props.popoverClassName)}
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
