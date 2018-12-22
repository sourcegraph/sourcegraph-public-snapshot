import React from 'react'
import { NOOP_TELEMETRY_SERVICE, TelemetryService } from './telemetryService'

/**
 * A React context that holds the telemetry service (for logging telemetry events).
 */
export const TelemetryContext = React.createContext<TelemetryService>(NOOP_TELEMETRY_SERVICE)
