import { ErrorInfo } from 'react'

/**
 * Ensures consistent `ErrorContext` across `addError` calls.
 * https://docs.datadoghq.com/real_user_monitoring/browser/modifying_data_and_context/?tab=npm#enrich-and-control-rum-data
 */
interface ErrorContext {
    // Used to ignore some errors in the `beforeSend` hook.
    originalException: unknown
    errorInfo?: ErrorInfo
}

export function isDatadogRumAvailable(): boolean {
    return typeof DD_RUM !== 'undefined' && typeof DD_RUM.addError !== 'undefined'
}

export const DatadogClient = {
    addError: (error: unknown, context: ErrorContext): void => {
        // Temporary solution for checking the availability of the
        // Datadog SDK until we decide to move forward with this service.
        if (isDatadogRumAvailable()) {
            DD_RUM.addError(error, context)
        }
    },
}
