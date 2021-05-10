import {
    EventHandler,
    FormEventHandler,
    RefObject,
    SyntheticEvent,
    useEffect,
    useRef,
    useState,
} from 'react'
import { noop } from 'rxjs';

// Special key for the submit error store.
export const FORM_ERROR = 'useForm/submissionErrors'

export type SubmissionErrors = Record<string, any> | undefined

interface UseFormProps<FormValues extends object> {
    /**
     * Initial values for form fields.
     * */
    initialValues: FormValues

    /**
     * Submit handler for a form element.
     * */
    onSubmit: (values: FormValues) => SubmissionErrors | Promise<SubmissionErrors> | void

    /**
     * Change handler will be called every time when some field withing the form
     * has been changed with last fields values.
     * */
    onChange?: (values: FormValues) => void
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
    initialValues: FormValues

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
    value: Value

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
 * Store object of all fields state within the form element.
 * Used below to keep tracking of fields value, touched, validity and other
 * fields state data.
 * */
type FieldsState<FormValues> = Record<keyof FormValues, FieldState<unknown>>;

/**
 * Unified form abstraction to track form state and provide form fields management
 * React hook to have all needed state for building proper UX for forms.
 *
 * useForm is one of two hooks for form management which responsible for
 * form state - submitted, submitting, state of all form fileds from useField
 * hook.
 * */
export function useForm<FormValues extends object>(props: UseFormProps<FormValues>): Form<FormValues> {
    const {
        onSubmit,
        initialValues,
        onChange = noop } = props

    const [submitted, setSubmitted] = useState(false)
    const [submitting, setSubmitting] = useState(false)
    const [submitErrors, setSubmitErrors] = useState<SubmissionErrors>()
    const [fields, setFields] = useState<FieldsState<FormValues>>({} as FieldsState<FormValues>)

    const formElementReference = useRef<HTMLFormElement>(null)
    const onSubmitReference = useRef<UseFormProps<FormValues>['onSubmit']>()

    // Track unmounted state to prevent setState if async validation or async submitting
    // will be resolved after component has been unmounted.
    const isUnmounted = useRef<boolean>(false)

    const setFieldState = (name: keyof FormValues, state: FieldState<unknown>): void => {

        setFields(fields => ({ ...fields, [name]: state }))

        // On first render all fields within the form trigger setFieldState
        // in order to set initial state. OnChange handler shouldn't being run
        // on first render so we have to skip setFieldState calls during the first
        // fields render.
        const hasRegisterField = fields[name] !== undefined;

        if (hasRegisterField && fields[name].value !== state.value) {
            onChange(getFormValues({ ...fields, [name]: state }))
        }
    }

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
                setSubmitting(true)

                const submitResult = await onSubmitReference.current?.(getFormValues(fields))

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

/**
 * Creates form values object and omit all other internal state of form field.
 * Used to form values for onSubmit and onChange handlers.
 * */
function getFormValues<FormValues>(fields: Record<string, FieldState<unknown>>): FormValues {
    return (Object.keys(fields)).reduce(
        (values, fieldName) => ({ ...values, [fieldName]: fields[fieldName].value }),
        {} as FormValues
    )
}
