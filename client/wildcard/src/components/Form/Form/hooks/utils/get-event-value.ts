import type { ChangeEvent } from 'react'

/**
 * Type guard for change event. Since useField might be used on custom element there's the case
 * when onChange handler will be called on custom element without synthetic event but with some
 * custom input value.
 * */
function isChangeEvent<Value>(possibleEvent: ChangeEvent | Value): possibleEvent is ChangeEvent {
    return !!(possibleEvent as ChangeEvent).target
}

/**
 * Selector function which takes target value from the event.
 *
 * We can have a few different source of value due to what kind of event
 * we've got. For example: Checkbox - target.checked, input element - target.value
 * and if run onChange on custom form field therefore we've got value as event itself.
 * */
export function getEventValue<Value>(event: ChangeEvent<HTMLInputElement> | Value): Value {
    if (isChangeEvent(event)) {
        // Checkbox input case
        if (event.target.type === 'checkbox') {
            return event.target.checked as unknown as Value
        }

        // Native input value case
        return event.target.value as unknown as Value
    }

    // Custom input without event but with value of input itself.
    return event
}
