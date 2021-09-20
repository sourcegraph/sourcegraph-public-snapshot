import { ConfiguredRegistryExtension } from '@sourcegraph/shared/src/extensions/extension'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../../auth'
import { RouteDescriptor } from '../../util/contributions'

export interface ExtensionAreaRoute extends RouteDescriptor<ExtensionAreaRouteContext> {}

/**
 * Properties passed to all page components in the registry extension area.
 */
export interface ExtensionAreaRouteContext
    extends SettingsCascadeProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps {
    /** The extension registry area main URL. */
    url: string

    /** The extension that is the subject of the page. */
    extension: ConfiguredRegistryExtension<GQL.IRegistryExtension>

    onDidUpdateExtension: () => void

    /** The currently authenticated user. */
    authenticatedUser: AuthenticatedUser | null
}
