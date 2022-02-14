import { SettingsCascadeProps } from '@sourcegraph/client-api'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

/**
 * Common props for extend point component for rendering extension views section.
 */
export interface ExtensionViewsSectionCommonProps
    extends SettingsCascadeProps<Settings>,
        ExtensionsControllerProps<'extHostAPI'>,
        PlatformContextProps<'updateSettings'>,
        TelemetryProps {
    className?: string
}
