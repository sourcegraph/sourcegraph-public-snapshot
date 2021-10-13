import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

/**
 * Common props for extend point component for rendering extension views section.
 */
export interface ExtensionViewsSectionCommonProps extends ExtensionsControllerProps<'extHostAPI'>, TelemetryProps {
    className?: string
}
