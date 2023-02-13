import {
    ChangeEvent,
    Dispatch,
    InputHTMLAttributes,
    RefObject,
    useCallback,
    useLayoutEffect,
    useRef,
    useState,
} from 'react'

import { noop } from 'rxjs'

import { FieldMetaState, FieldState, FormAPI } from './useForm'
import { getEventValue } from './utils/get-event-value'
import { useAsyncValidation } from './utils/use-async-validation'
import { AsyncValidator, getCustomValidationContext, getCustomValidationMessage, Validator } from './validators'

export interface Validators<FieldValue, ErrorContext> {
    sync?: Validator<FieldValue, ErrorContext>
    async?: AsyncValidator<FieldValue>
}

/**
 * Subset of native input props that useField can set to the native input element.
 */
export interface InputProps<Value>
    extends Omit<InputHTMLAttributes<HTMLInputElement>, 'name' | 'value' | 'onChange' | 'onBlur'> {
    onBlur?: () => void
    onChange?: (value: Value) => void
}

/**
 * Public API for input element. Contains all handlers and props for
 * native input element and expose meta state of input like touched,
 * validState etc.
 */
export interface useFieldAPI<FieldValue, ErrorContext = unknown> {
    /**
     * Props and handles which should be passed to input component in order
     * to track change of value, set controlled value and set errors by
     * Constraint validation API.
     */
    input: {
        ref: RefObject<HTMLInputElement & HTMLFieldSetElement>
        name: string
        value: FieldValue
        onChange: (event: ChangeEvent<HTMLInputElement> | FieldValue) => void
        onBlur: () => void
    } & InputProps<FieldValue>

    /**
     * Meta state of form field - like touched, valid state and last
     * native validity state.
     */
    meta: FieldState<FieldValue> & {
        validationContext: ErrorContext | undefined
        /**
         * Set state handler gives access to set inner state of useField.
         * Useful for complicated cases when need to deal with custom react input
         * component.
         */
        setState: (dispatch: (previousState: FieldState<FieldValue>) => FieldState<FieldValue>) => void
    }
}

export type UseFieldProps<FormValues, Key, Value, ErrorContext> = {
    name: Key
    formApi: FormAPI<FormValues>
    validators?: Validators<Value, ErrorContext>
} & InputProps<Value>

/**
 * React hook to manage validation of a single form input field.
 * `useInputValidation` helps with coordinating the constraint validation API
 * and custom synchronous and asynchronous validators.
 *
 * Should be used with useForm hook to connect field and form component's states.
 */
export function useField<ErrorContext, FormValues, Key extends keyof FormValues>(
    props: UseFieldProps<FormValues, Key, FormValues[Key], ErrorContext>
): useFieldAPI<FormValues[Key], ErrorContext> {
    const { formApi, name, validators, onChange = noop, disabled, ...inputProps } = props
    const { submitted, touched: formTouched } = formApi
    const { sync: syncValidator, async: asyncValidator } = validators ?? {}

    const inputReference = useRef<HTMLInputElement & HTMLFieldSetElement>(null)

    const [state, setState] = useFormFieldState(name, formApi)
    const { start: startAsyncValidation, cancel: cancelAsyncValidation } = useAsyncValidation({
        inputReference,
        asyncValidator,
        onValidationChange: asyncState => setState(previousState => ({ ...previousState, ...asyncState })),
    })

    // Since validation logic wants to use sync state update we use `useLayoutEffect` instead of
    // `useEffect` in order to synchronously re-render after value setState updates, but before
    // the browser has painted DOM updates. This prevents users from seeing inconsistent states
    // where changes handled by React have been painted, but DOM manipulation handled by these
    // effects are painted on the next tick.
    useLayoutEffect(() => {
        const inputElement = inputReference.current

        // Clear custom validity from the last validation call.
        inputElement?.setCustomValidity?.('')

        const validity = inputElement?.validity ?? null

        // If we got error from native attr validation (required, pattern, type)
        // we still run validator in order to get some custom error message for
        // standard validation error if validator doesn't provide message we fall back
        // on standard validationMessage string [1] (ex. Please fill in input.)
        const nativeErrorMessage = inputElement?.validationMessage ?? ''
        const customValidationResult = syncValidator ? syncValidator(state.value, validity) : undefined

        const customValidationMessage = getCustomValidationMessage(customValidationResult)
        const customValidationContext = getCustomValidationContext(customValidationResult)

        if (customValidationMessage || (!customValidationResult && nativeErrorMessage)) {
            // We have to cancel async validation from previous call
            // if we got sync validation native or custom.
            cancelAsyncValidation()

            // [1] Custom error message or fallback on native error message
            const validationMessage = customValidationMessage || nativeErrorMessage

            inputElement?.setCustomValidity?.(validationMessage)

            return setState(state => ({
                ...state,
                validState: 'INVALID',
                error: validationMessage,
                errorContext: customValidationContext,
                validity,
            }))
        }

        if (asyncValidator) {
            // Due to the call of start async validation in useLayoutEffect we have to
            // schedule the async validation event in the next tick to be able to run
            // observable pipeline validation since useAsyncValidation hook use
            // useObservable hook internally which calls '.subscribe' in useEffect.
            requestAnimationFrame(() => {
                startAsyncValidation({ value: state.value, validity })
            })

            return setState(state => ({
                ...state,
                validState: 'CHECKING' as const,
                error: '',
                validity,
            }))
        }

        // Hide any validation errors by default if the field is in the disabled
        // state.
        if (disabled && !syncValidator && !asyncValidator) {
            return setState(state => ({
                ...state,
                validState: 'NOT_VALIDATED',
                error: '',
                validity,
            }))
        }

        return setState(state => ({
            ...state,
            validState: 'VALID' as const,
            error: '',
            errorContext: customValidationContext,
            validity,
        }))
    }, [state.value, syncValidator, startAsyncValidation, asyncValidator, cancelAsyncValidation, setState, disabled])

    const handleBlur = useCallback(() => setState(state => ({ ...state, touched: true })), [setState])
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement> | FormValues[Key]) => {
            const value = getEventValue(event)

            setState(state => ({ ...state, value, dirty: true }))
            onChange(value)
        },
        [onChange, setState]
    )

    return {
        input: {
            name: name.toString(),
            ref: inputReference,
            value: state.value,
            onBlur: handleBlur,
            onChange: handleChange,
            disabled,
            ...inputProps,
        },
        meta: {
            ...state,
            validationContext: state.errorContext as ErrorContext,
            touched: state.touched || submitted || formTouched,
            // Set state dispatcher gives access to set inner state of useField.
            // Useful for complex cases when you need to deal with custom react input
            // components.
            setState: dispatch => {
                setState(state => ({ ...dispatch(state) }))
            },
        },
    }
}

type FieldStateTransformer<Value> = (previousState: FieldState<Value>) => FieldState<Value>

function useFormFieldState<FormValues, Key extends keyof FormValues>(
    name: Key,
    formAPI: FormAPI<FormValues>
): [FieldState<FormValues[Key]>, Dispatch<FieldStateTransformer<FormValues[Key]>>] {
    const { fields, setFieldState } = formAPI
    const state = fields[name]

    // Use useRef for form api handler in order to avoid unnecessary
    // calls if API handler has been changed.
    const setFieldStateReference = useRef(setFieldState)
    setFieldStateReference.current = setFieldState

    const setState = useCallback(
        (transformer: FieldStateTransformer<FormValues[Key]>) => {
            setFieldStateReference.current(name, transformer as FieldStateTransformer<unknown>)
        },
        [name]
    )

    return [state, setState]
}

export type UseControlledFieldProps<Value> = {
    value: Value
    name: string
    submitted: boolean
    formTouched: boolean
    validators?: Validators<Value, unknown>
} & InputProps<Value>

/**
 * React hook to manage validation of a single form input field.
 * `useInputValidation` helps with coordinating the constraint validation API
 * and custom synchronous and asynchronous validators.
 *
 * Should be used with useForm hook to connect field and form component's states.
 */
export function useControlledField<Value>(props: UseControlledFieldProps<Value>): useFieldAPI<Value, unknown> {
    const { value, validators, submitted, formTouched, onChange = noop, name, disabled, ...inputProps } = props
    const { sync: syncValidator, async: asyncValidator } = validators ?? {}

    const inputReference = useRef<HTMLInputElement & HTMLFieldSetElement>(null)
    const [state, setState] = useState<FieldMetaState<Value>>({
        touched: false,
        dirty: false,
        validState: 'NOT_VALIDATED',
        validity: null,
        initialValue: value,
    })

    const { start: startAsyncValidation, cancel: cancelAsyncValidation } = useAsyncValidation({
        inputReference,
        asyncValidator,
        onValidationChange: asyncState => setState(previousState => ({ ...previousState, ...asyncState })),
    })

    useLayoutEffect(() => {
        const inputElement = inputReference.current

        // Clear custom validity from the last validation call.
        inputElement?.setCustomValidity?.('')

        const validity = inputElement?.validity ?? null

        // If we got error from native attr validation (required, pattern, type)
        // we still run validator in order to get some custom error message for
        // standard validation error if validator doesn't provide message we fall back
        // on standard validationMessage string [1] (ex. Please fill in input.)
        const nativeErrorMessage = inputElement?.validationMessage ?? ''
        const customValidationResult = syncValidator ? syncValidator(value, validity) : undefined

        const customValidationMessage = getCustomValidationMessage(customValidationResult)
        const customValidationContext = getCustomValidationContext(customValidationResult)

        if (customValidationMessage || (!customValidationResult && nativeErrorMessage)) {
            // We have to cancel async validation from previous call
            // if we got sync validation native or custom.
            cancelAsyncValidation()

            // [1] Custom error message or fallback on native error message
            const validationMessage = customValidationMessage || nativeErrorMessage

            inputElement?.setCustomValidity?.(validationMessage)

            return setState(state => ({
                ...state,
                validState: 'INVALID',
                error: validationMessage,
                errorContext: customValidationContext,
                validity,
            }))
        }

        if (asyncValidator) {
            // Due to the call of start async validation in useLayoutEffect we have to
            // schedule the async validation event in the next tick to be able to run
            // observable pipeline validation since useAsyncValidation hook use
            // useObservable hook internally which calls '.subscribe' in useEffect.
            requestAnimationFrame(() => {
                startAsyncValidation({ value, validity })
            })

            return setState(state => ({
                ...state,
                validState: 'CHECKING' as const,
                error: '',
                validity,
            }))
        }

        // Hide any validation errors by default if the field is in the disabled
        // state.
        if (disabled && !syncValidator && !asyncValidator) {
            return setState(state => ({
                ...state,
                validState: 'NOT_VALIDATED',
                error: '',
                validity,
            }))
        }

        return setState(state => ({
            ...state,
            validState: 'VALID' as const,
            error: '',
            errorContext: customValidationContext,
            validity,
        }))
    }, [value, syncValidator, startAsyncValidation, asyncValidator, cancelAsyncValidation, setState, disabled])

    const handleBlur = useCallback(() => setState(state => ({ ...state, touched: true })), [setState])
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement> | Value) => {
            const value = getEventValue(event)

            onChange(value)
            setState(state => ({ ...state, dirty: true }))
        },
        [onChange, setState]
    )

    return {
        input: {
            value,
            name: name.toString(),
            ref: inputReference,
            onBlur: handleBlur,
            onChange: handleChange,
            disabled,
            ...inputProps,
        },
        meta: {
            ...state,
            value,
            validationContext: state.errorContext,
            touched: state.touched || submitted || formTouched,
            // Set state dispatcher gives access to set inner state of useField.
            // Useful for complex cases when you need to deal with custom react input
            // components.
            setState: dispatch => {
                setState(state => ({ ...dispatch({ value, ...state }) }))
            },
        },
    }
}
