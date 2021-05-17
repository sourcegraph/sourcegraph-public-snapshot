import { ChangeEvent, FocusEventHandler, RefObject, useEffect, useRef, useState } from 'react'
import { noop } from 'rxjs'

import { FieldState, FormAPI, ValidationResult } from './useForm'
import { getEventValue } from './utils/get-event-value';
import { AsyncValidator, useAsyncValidation } from './utils/use-async-validation';

export type Validator<FieldValue> = (value: FieldValue | undefined, validity?: ValidityState | null) => ValidationResult

export interface Validators<FieldValue> {
    sync?: Validator<FieldValue>,
    async?: AsyncValidator<FieldValue>
}

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
    validators?: Validators<FormValues[FieldValueKey]>,
): useFieldAPI<FormValues[FieldValueKey]> {
    const { setFieldState, initialValues, submitted, touched: formTouched } = formApi

    const { sync = noop, async } = validators ?? {};

    const inputReference = useRef<HTMLInputElement & HTMLFieldSetElement>(null)

    const [selfValue, setSelfValue] = useState<FormValues[FieldValueKey]>(initialValues[name])

    const [state, setState] = useState<FieldState<FormValues[FieldValueKey]>>({
        value: selfValue,
        touched: false,
        validState: 'NOT_VALIDATED',
        error: '',
        validity: null,
    })

    const { start: startAsyncValidation, cancel: cancelAsyncValidation } = useAsyncValidation({
        inputReference,
        asyncValidator: async,
        onValidationChange: state => setState(
            previousState => ({ ...previousState, ...state, value: selfValue })
        )
    })

    // Use useRef for form api handler in order to avoid unnecessary
    // calls if API handler has been changed.
    const setFieldStateReference = useRef<FormAPI<FormValues>['setFieldState']>(setFieldState)
    setFieldStateReference.current = setFieldState

    useEffect(() => {
        console.log('HELLO use field validity effect')
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
        const customValidation = sync(selfValue, validity)

        if (customValidation || !nativeAttributeValidation) {
            // We have to cancel async validation from previous call
            // if we got sync validation native or custom.
            cancelAsyncValidation()

            // [1] Custom error message or fallback on native error message
            const validationMessage = customValidation || nativeErrorMessage

            inputElement?.setCustomValidity?.(validationMessage)

            return setState(state => ({
                ...state,
                value: selfValue,
                validState: 'INVALID',
                error: validationMessage,
                validity,
            }))
        }

        if (async) {
            startAsyncValidation({ value: selfValue, validity })
        }

        return setState(state => ({
            ...state,
            value: selfValue,
            validState: 'VALID' as const,
            error: '',
            validity,
        }))
    }, [selfValue, sync, startAsyncValidation, async, cancelAsyncValidation])

    // Sync field state with state on form level - useForm hook will used this state to run
    // onSubmit handler and track validation state to prevent onSubmit run when async
    // validation is going.
    useEffect(() => setFieldStateReference.current(name, state), [name, state])

    return {
        input: {
            name: name.toString(),
            ref: inputReference,
            value: selfValue,
            onBlur: () => setState(state => ({ ...state, touched: true })),
            onChange: (event: ChangeEvent<HTMLInputElement> | FormValues[FieldValueKey]) => {
                const value = getEventValue(event)

                setSelfValue(value)
            },
        },
        meta: {
            ...state,
            value: selfValue,
            touched: state.touched || submitted || formTouched,
        },
    }
}

