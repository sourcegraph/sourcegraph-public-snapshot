import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

/**
 * Common props for extend point component for rendering extension views section.
 */
export interface ExtensionViewsSectionCommonProps
    extends SettingsCascadeProps<Settings>,
        ExtensionsControllerProps<'extHostAPI'>,
        TelemetryProps {
    className?: string
}
