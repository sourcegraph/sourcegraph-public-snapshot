/**
 * EXPERIMENTAL
 *
 * EventName enumerates all known metadata keys. These must be statically
 * defined up-front to avoid risks of accidental leakage of private instance
 * data.
 *
 * Do NOT forcibly cast an arbitrary string to this type, except with the
 * modifiers declared in this package.
 */
export enum EventName {
    FooBar = 'FooBar',
    hover = 'hover',
    UserEventLogPage = 'UserEventLogPage',
}

/**
 * EXPERIMENTAL
 *
 * This EventName namespace extends EventName with some event name modifiers
 * that we support. This is the only place where values can be cast to EventName.
 */
export namespace EventName {
    /**
     * View prefixes eventName with a modifier for indicating a page view event.
     * The format is `View${eventName}`
     */
    export function View(eventName: EventName): EventName {
        return `View${eventName}` as EventName
    }
}

/**
 * EXPERIMENTAL
 *
 * MetadataKey enumerates all known metadata keys. These must be statically
 * defined up-front to avoid risks of accidental leakage of private instance
 * data.
 *
 * Do NOT forcibly cast an arbitrary string to this type.
 */
export enum MetadataKey {
    Foo = 'Foo',
}

/**
 * EXPERIMENTAL
 *
 * Props interface that can be extended by React components depending on the
 * TelemetryV2Service.
 */
export interface TelemetryPropsV2 {
    /**
     * A telemetry service implementation to log events.
     */
    telemetryServiceV2: TelemetryServiceV2
}

/**
 * EXPERIMENTAL
 *
 * EventParameters describes additional, optional parameters for recording events.
 */
export type EventParameters = {
    /**
     * version should indicate the version of the shape of this particular
     * event.
     */
    version?: number
    /**
     * metadata is array of tuples with predefined keys and arbitrary
     * numeric value. This data is always exported alongside events to
     * Sourcegraph.
     *
     * Typescript has poor support for excess property checking on objects,
     * so this is the easiest way to enforce that keys belong to statically
     * defined enums.
     */
    metadata?: [[MetadataKey, number]]
    /**
     * privateMetadata is an object with arbitrary keys and values. This
     * is NOT exported by default, as it may contain private instance data.
     */
    privateMetadata?: { [key: string]: string }
}

/**
 * The telemetry service records events for forwarding to Sourcegraph.
 *
 * EXPERIMENTAL
 */
export interface TelemetryServiceV2 {
    /**
     * Record an event (by sending it to the server, which forwards it to
     * Sourcegraph).
     */
    record(name: EventName, parameters?: EventParameters): void

    recordString(name: string, parameters?: EventParameters): void
}
