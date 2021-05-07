import {
    EventHandler,
    FormEventHandler,
    RefObject,
    SyntheticEvent,
    useCallback,
    useEffect,
    useRef,
    useState,
} from 'react'

// Special key for the submit error store.
export const FORM_ERROR = 'useForm/submissionErrors'

export type SubmissionErrors = Record<string, any> | undefined

interface UseFormProps<FormValues extends object> {
    /**
     * Initial values for form fields.
     * */
    initialValues: Partial<FormValues>

    /**
     * Submit handlers for a form element.
     * */
    onSubmit: (values: FormValues) => SubmissionErrors | Promise<SubmissionErrors> | void
}

/**
 * High level API for form instance. It consists with form api for useField hook,
 * form state, form handlers like handleSubmit and some props which should be
 * passed on root form element.
 *
 * */
export interface Form<FormValues> {
    /**
     * State and methods of form, used in consumers to create filed by useField(formAPI)
     * */
    formAPI: FormAPI<FormValues>

    /**
     * Handler for onSubmit form element.
     * */
    handleSubmit: FormEventHandler | EventHandler<SyntheticEvent>

    /**
     * Ref for the root element of form. It might be form or any html element.
     * Used to find first invalid input within this root element and call focus
     * on it.
     * */
    ref: RefObject<any>
}

/**
 * This API should be passed to useField hook for registration they state
 * to the form object from useForm hook. Also this api consists form state
 * like submitting, submitted, validation.
 * */
export interface FormAPI<FormValues> {
    /**
     * Initial values for the form.
     * These values are set for inputs only in first render.
     * Initial values also used as field value for first run
     * of sync and async validators on useField level.
     * */
    initialValues: Partial<FormValues>

    /**
     * Mark to understand was there an attempt by user to submit the form?
     * Used in useField hook to trigger appearance of error message if
     * user tried submit the form.
     * */
    submitted: boolean

    /**
     * State to understand there some field is processing async validations.
     * It might be useful to disable submit button for example if we got async
     * validation for form filed.
     * */
    validating: boolean

    /**
     * State to understand that form submitting is going on.
     * Also might be used as a sign to disable or show loading
     * state for submit button.
     * */
    submitting: boolean

    /** Store for submit errors which we got from onSubmit prop handler */
    submitErrors: SubmissionErrors

    /**
     * Public api for register fields to the form from useField hook.
     * By this we have field state withing useField hook and in useField.
     * */
    setFieldState: (name: keyof FormValues, state: FieldState<unknown>) => void
}

/**
 * Field state which present public state from useField hook. On order to aggregate
 * state of all fields within the form we store all fields state on form level as well.
 * */
export interface FieldState<Value> {
    /**
     * Field (input) controlled value. This value might be not only some primitive value
     * like string, number but array, object, tuple and other complex types as consumer set.
     * */
    value: Value | undefined

    /**
     * State to understand when users focused and blurred input element.
     * */
    touched: boolean

    /**
     * Valid state with initial value NOT_VALIDATED, with VALID when all validators
     * didn't return validation error, CHECKING for when async validation is going on,
     * and INVALID when some validator returns validation error.
     * */
    validState: 'VALID' | 'INVALID' | 'NOT_VALIDATED' | 'CHECKING'

    /**
     * Last error value which has been returned from validators.
     * */
    error?: any

    /**
     * Native validity state from native validation API of input element.
     * Null when useField is used for some custom elements instead of native input.
     * */
    validity: ValidityState | null
}

/**
 * Unified form abstraction to track form state and provide form fields management
 * React hook to have all needed state for building proper UX for forms.
 *
 * useForm is one of two hooks for form management which responsible for
 * form state - submitted, submitting, state of all form fileds from useField
 * hook.
 * */
export function useForm<FormValues extends object>(props: UseFormProps<FormValues>): Form<FormValues> {
    const { onSubmit, initialValues } = props

    const [submitted, setSubmitted] = useState(false)
    const [submitting, setSubmitting] = useState(false)
    const [submitErrors, setSubmitErrors] = useState<SubmissionErrors>()
    const [fields, setFields] = useState<Record<string, FieldState<unknown>>>({})

    const formElementReference = useRef<HTMLFormElement>(null)
    const onSubmitReference = useRef<UseFormProps<FormValues>['onSubmit']>()

    // Track unmounted state to prevent setState if async validation or async submitting
    // will be resolved after component has been unmounted.
    const isUnmounted = useRef<boolean>(false)

    const setFieldState = useCallback((name: keyof FormValues, state: FieldState<unknown>) => {
        setFields(fields => ({ ...fields, [name]: state }))
    }, [])

    useEffect(
        () => () => {
            isUnmounted.current = true
        },
        []
    )

    // Mutate local ref for submit handler to allow pass onSubmit
    // handler without memo.
    onSubmitReference.current = onSubmit

    return {
        formAPI: {
            submitted,
            submitting,
            submitErrors,
            initialValues,
            setFieldState,
            validating: Object.values<FieldState<unknown>>(fields).some(state => state.validState === 'CHECKING'),
        },
        ref: formElementReference,
        handleSubmit: async event => {
            event.preventDefault()

            setSubmitted(true)

            const hasInvalidField = Object.values<FieldState<unknown>>(fields).some(
                state => state.validState === 'INVALID'
            )

            if (!hasInvalidField) {
                // Collect all form fields to pass them to onSubmit handler.
                const values = Object.keys(fields).reduce<FormValues>(
                    (values, fieldName) => ({ ...values, [fieldName]: fields[fieldName].value }),
                    {} as FormValues
                )

                setSubmitting(true)

                const submitResult = await onSubmitReference.current?.(values)

                // Check isUnmounted state to prevent calling setState on
                // unmounted components.
                if (!isUnmounted.current) {
                    setSubmitting(false)
                    // eslint-disable-next-line no-unused-expressions
                    submitResult && setSubmitErrors(submitResult)
                }
            } else {
                // Hack to focus first invalid input on submit, since we are not using
                // native behavior in order to avoid poor UX of native validation focus on error
                // we have to find and focus invalid input by ourselves
                formElementReference.current?.querySelector<HTMLInputElement>('input:invalid')?.focus()
            }
        },
    }
}
