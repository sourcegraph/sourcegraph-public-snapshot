/* eslint-disable @typescript-eslint/consistent-type-definitions */
declare module 'libhoney' {
    type LibhoneyOptions = {
        writeKey: string
        dataset: string
    }

    type FieldValue = string | number | undefined | null | boolean
    type DynamicFieldValue = () => FieldValue

    type Fields = Record<string, FieldValue>
    type DynamicFields = Record<string, DynamicFieldValue>

    /**
     * Represents an individual event to send to Honeycomb.
     */
    class Event {
        public timestamp: string
        public data: Fields

        constructor(libhoney: Libhoney, fields: Fields, dynamicFields: DynamicFields)

        /**
         * adds a group of field->values to this event.
         */
        add(data: Fields): void

        /**
         * adds a single field->value mapping to this event.
         */
        addField(name: string, value: FieldValue): Event

        /**
         * attaches data to an event that is not transmitted to honeycomb, but instead is available
         * when checking the send responses.
         */
        addMetadata(md: Record<string, string>): Event

        /**
         * Sends this event to honeycomb, sampling if necessary.
         */
        send(): void

        /**
         * Dispatch an event to be sent to Honeycomb.  Assumes sampling has already happened,
         * and will send every event handed to it.
         */
        sendPresampled(): void
    }

    /*
     * Allows piecemeal creation of events.
     */
    class Builder {
        constructor(libhoney: Libhoney, fields: Fields, dynamicFields: DynamicFields)
        /*
         * adds a group of field->values to the events created from this builder.
         */
        add(fields: Fields & DynamicFields): Builder
        /*
         * adds a single field->value mapping to the events created from this builder.
         */
        addField(name: string, value: FieldValue): Builder
        /*
     * adds a single field->dynamic value function, which is invoked to supply values when events
    are created from this builder.
     */
        addDynamicField(name: string, dynamicFieldValue: DynamicFieldValue): Builder
        /**
         * creates and sends an event, including all builder fields/dynFields, as well as anything
         * in the optional data parameter.
         */
        sendNow(fields: Fields): void
        /*
     * creates and returns a new Event containing all fields/dynFields from this builder, that
    can be further fleshed out and sent on its own.
     */
        newEvent(): Event
        /*
         * creates and returns a clone of this builder, merged with fields and dynFields passed as
         * arguments.
         */
        newBuilder(fields: Fields, dynamicFields?: DynamicFields): Builder
    }

    /*
     * libhoney aims to make it as easy as possible to create events and send them on into Honeycomb.
     */
    class Libhoney {
        public sampleRate: number
        public dataset: number
        public writeKey: string
        public apiHost: string
        /*
         * Constructs a libhoney context in order to configure default behavior,
         * though each of its members (`apiHost`, `writeKey`, `dataset`, and
         * `sampleRate`) may in fact be overridden on a specific Builder or Event.
         */
        constructor(options: LibhoneyOptions)
        add: Builder['add']
        addField: Builder['addField']
        addDynamicField: Builder['addDynamicField']
        sendNow: Builder['sendNow']
        newEvent: Builder['newEvent']
        newBuilder: Builder['newBuilder']
        /**
         * Allows you to easily wait for everything to be sent to Honeycomb (and for responses to come back for
         * events). Also initializes a transmission instance for libhoney to use, so any events sent
         * after a call to flush will not be waited on.
         */
        flush(): Promise<void>
    }

    export default Libhoney
}
