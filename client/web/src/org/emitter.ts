let _bus: Comment
type EventBusUnsubscribeHandle = (event: CustomEvent) => void
type EventBusEvents = 'refreshOrgHeader'

export interface IEventBus {
    subscribe: <T>(
        event: EventBusEvents,
        callback: (data?: T) => void,
        options?: AddEventListenerOptions
    ) => EventBusUnsubscribeHandle

    unSubscribe: (
        event: EventBusEvents,
        unsubscribeHandle: EventBusUnsubscribeHandle,
        options?: AddEventListenerOptions
    ) => void

    emit: <T>(type: EventBusEvents, data: T) => void
}

const subscribe = <T>(
    event: EventBusEvents,
    callback: (data?: T) => void,
    options?: AddEventListenerOptions
): EventBusUnsubscribeHandle => {
    const callBack = (event: CustomEvent<T>): void => {
        const eventData = event.detail
        callback(eventData)
    }
    _bus.addEventListener(event, callBack as EventListenerOrEventListenerObject, options)
    return callBack
}

const unSubscribe = (
    event: EventBusEvents,
    unsubscribeHandle: EventBusUnsubscribeHandle,
    options?: AddEventListenerOptions
): void => {
    _bus.removeEventListener(event, unsubscribeHandle as EventListenerOrEventListenerObject, options)
}

const emit = <T>(type: EventBusEvents, data: T): void => {
    const event = new CustomEvent(type, {
        detail: data,
    })
    _bus.dispatchEvent(event)
}

export const useEventBus = (): IEventBus => {
    if (!_bus) {
        _bus = new Comment('OrgArea-event-bus')
        document.append(_bus)
    }

    return {
        subscribe,
        unSubscribe,
        emit,
    }
}
