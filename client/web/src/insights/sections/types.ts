import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../schema/settings.schema'

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
