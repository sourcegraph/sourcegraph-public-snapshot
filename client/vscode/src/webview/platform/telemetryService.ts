import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

const vscodeTelemetryService: TelemetryService = {
    // TODO: generate and store anon user id.
    // store w Memento

    log: () => {},
    logViewEvent: () => {},
}
