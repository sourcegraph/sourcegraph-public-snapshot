import { ChangeEvent, FocusEventHandler, RefObject, useEffect, useRef, useState } from 'react'
import { noop } from 'rxjs'

import { FieldState, FormAPI } from './useForm'

export type ValidationResult = string | undefined | void
export type Validator<FieldValue> = (value: FieldValue | undefined, validity: ValidityState | null) => ValidationResult

/**
 * Public API for input element. Contains all handlers and props for
 * native input element and expose meta state of input like touched,
 * validState and etc.
 * */
export interface useFieldAPI<FieldValue> {
    /**
     * Props and handles which should be passed to input component in order
     * to track change of value, set controlled value and set errors by
     * Constraint validation API.
     * */
    input: {
        ref: RefObject<HTMLInputElement & HTMLFieldSetElement>
        name: string
        value: FieldValue | undefined
        onChange: (event: ChangeEvent<HTMLInputElement> | FieldValue) => void
        onBlur: FocusEventHandler<HTMLInputElement>
    }
    /**
     * Meta state of form field - like touched, valid state and last
     * native validity state.
     * */
    meta: FieldState<FieldValue>
}

/**
 * React hook to manage validation of a single form input field.
 * `useInputValidation` helps with coordinating the constraint validation API
 * and custom synchronous and asynchronous validators.
 *
 * Should be used with useForm hook to connect field and form component's states.
 * */
export function useField<FormValues, FieldValueKey extends keyof FormAPI<FormValues>['initialValues']>(
    name: FieldValueKey,
    formApi: FormAPI<FormValues>,
    validator: Validator<FormValues[FieldValueKey]> = noop
): useFieldAPI<FormValues[FieldValueKey]> {
    const { setFieldState, initialValues, submitted, touched: formTouched } = formApi

    const inputReference = useRef<HTMLInputElement & HTMLFieldSetElement>(null)
    const [state, setState] = useState<FieldState<FormValues[FieldValueKey]>>({
        value: initialValues[name],
        touched: false,
        validState: 'NOT_VALIDATED',
        error: '',
        validity: null,
    })

    // Use useRef for form api handler in order to avoid unnecessary
    // calls if API handler has been changed.
    const setFieldStateReference = useRef<FormAPI<FormValues>['setFieldState']>(setFieldState)
    setFieldStateReference.current = setFieldState

    useEffect(() => {
        const inputElement = inputReference.current

        // Clear custom validity from the last validation call.
        inputElement?.setCustomValidity?.('')

        const nativeAttributeValidation = inputElement?.checkValidity?.() ?? true
        const validity = inputElement?.validity ?? null

        // If we got error from native attr validation (required, pattern, type)
        // we still run validator in order to get some custom error message for
        // standard validation error if validator doesn't provide message we fallback
        // on standard validationMessage string [1] (ex. Please fill in input.)
        const nativeErrorMessage = inputElement?.validationMessage ?? ''
        const customValidation = validator(state.value, validity)

        if (customValidation || !nativeAttributeValidation) {
            // [1] Custom error message or fallback on native error message
            const validationMessage = customValidation || nativeErrorMessage

            inputElement?.setCustomValidity?.(validationMessage)

            return setState(state => ({
                ...state,
                validState: 'INVALID',
                error: validationMessage,
                validity,
            }))
        }

        return setState(state => ({
            ...state,
            validState: 'VALID' as const,
            error: '',
            validity,
        }))
    }, [state.value, validator])

    // Sync field state with state on form level - useForm hook will used this state to run
    // onSubmit handler and track validation state to prevent onSubmit run when async
    // validation is going.
    useEffect(() => setFieldStateReference.current(name, state), [name, state])

    return {
        input: {
            name: name.toString(),
            ref: inputReference,
            value: state.value,
            onBlur: () => setState(state => ({ ...state, touched: true })),
            onChange: (event: ChangeEvent<HTMLInputElement> | FormValues[FieldValueKey]) => {
                const value = getEventValue(event)

                setState(state => ({ ...state, value }))
            },
        },
        meta: {
            ...state,
            touched: state.touched || submitted || formTouched,
        },
    }
}

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
function getEventValue<Value>(event: ChangeEvent<HTMLInputElement> | Value): Value {
    if (isChangeEvent(event)) {
        // Checkbox input case
        if (event.target.type === 'checkbox') {
            return (event.target.checked as unknown) as Value
        }

        // Native input value case
        return (event.target.value as unknown) as Value
    }

    // Custom input without event but with value of input itself.
    return event
}
