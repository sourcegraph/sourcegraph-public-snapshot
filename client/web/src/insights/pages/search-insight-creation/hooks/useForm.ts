import {
    ChangeEvent, EventHandler,
    FocusEventHandler,
    FormEventHandler,
    RefObject, SyntheticEvent,
    useCallback,
    useEffect,
    useRef,
    useState,
} from 'react';
import { noop } from 'rxjs';

// Special key for the submit error store.
export const FORM_ERROR = 'useForm/submissionErrors';

export interface AnyObject {
    [key: string]: any
}

export type SubmissionErrors = AnyObject | undefined

interface UseFormProps<FormValues extends object> {
    /** Initial values for form fields. */
    initialValues: Partial<FormValues>
    /** Submit handlers for the form element. */
    onSubmit: (values: FormValues) => SubmissionErrors | Promise<SubmissionErrors> | void
}

interface Form<FormValues> {
    /** Internal state and methods of form, used in consumers to create filed by useField(formAPI) */
    formAPI: FormAPI<FormValues>,
    /** Handler for onSubmit form element. */
    handleSubmit: FormEventHandler | EventHandler<SyntheticEvent>,
    /** Ref for the root element of form. It might be form or any html element. */
    ref: RefObject<any>
}

interface FormAPI<FormValues> {
    /** Initial values for the form. These values are set for inputs only in first render. */
    initialValues: Partial<FormValues>,
    /** Mark to understand was there an attempt to submit the form? */
    submitted: boolean,
    /** State for the sign that some field is processing async validation. */
    validating: boolean,
    /** Sign of form submitting is going on. */
    submitting: boolean,
    /** Store for submit errors. */
    submitErrors: SubmissionErrors
    /** Internal api for register field to the form. */
    setFieldState: (name: keyof FormValues, state: FieldState<unknown>) => void
}

/**
 * Unified form abstraction to track form state and form fields management
 * */
export function useForm<FormValues extends object>(props: UseFormProps<FormValues>): Form<FormValues> {
    const { onSubmit, initialValues } = props;

    const [submitted, setSubmitted] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [submitErrors, setSubmitErrors] = useState<SubmissionErrors>();
    const [fields, setFields] = useState<Record<string, FieldState<unknown>>>({})

    const formElementReference = useRef<HTMLFormElement>(null)
    const onSubmitReference = useRef<UseFormProps<FormValues>['onSubmit']>();
    const isUnmounted = useRef<boolean>(false);

    const setFieldState = useCallback((name: keyof FormValues, state: FieldState<unknown>) => {
        setFields(fields => ({...fields, [name]: state }))
    }, [])

    useEffect(() => () => { isUnmounted.current = true }, []);

    // Allow pass onSubmit
    onSubmitReference.current = onSubmit;

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
            event.preventDefault();

            setSubmitted(true);

            const hasInvalidField = Object.values<FieldState<unknown>>(fields).some(state => state.validState === 'INVALID');

            if (!hasInvalidField) {
                // hack to find and focus first invalid input withing the form.
                setSubmitting(true);

                const values = Object.keys(fields)
                    .reduce<FormValues>((values, fieldName) =>
                        ({...values, [fieldName]: fields[fieldName].value }),
                        {} as FormValues
                    )

                const submitResult = await onSubmitReference.current?.(values)

                if (!isUnmounted) {
                    setSubmitting(false);
                    // eslint-disable-next-line no-unused-expressions
                    submitResult && setSubmitErrors(submitResult);
                }
            } else {
                // Hack to focus first invalid input on submit, since we are not using
                // native behavior in order to avoid poor UX of native validation focus on error
                // we have to find and focus invalid input by ourselves
                formElementReference.current
                    ?.querySelector<HTMLInputElement>('input:invalid')
                    ?.focus();
            }
        },
    }
}

export type ValidationResult = string | undefined | void
export type Validator<FieldValue> = (value: FieldValue | undefined, validity: ValidityState | null) => ValidationResult

interface FieldState<Value> {
    value: Value | undefined,
    touched: boolean,
    validState: 'VALID' | 'INVALID' | 'NOT_VALIDATED' | 'CHECKING',
    error?: any
    validity: ValidityState | null
}

export interface useFieldAPI<FieldValue> {
    input: {
        ref: RefObject<HTMLInputElement & HTMLFieldSetElement>;
        name: string;
        value: FieldValue | undefined
        onChange: (event: ChangeEvent<HTMLInputElement> | FieldValue) => void
        onBlur: FocusEventHandler<HTMLInputElement>
    },
    meta: FieldState<FieldValue>
}

export function useField<FormValues, FieldValueKey extends keyof FormAPI<FormValues>['initialValues']>(
    name: FieldValueKey,
    formApi: FormAPI<FormValues>,
    validator: Validator<FormValues[FieldValueKey]> = noop): useFieldAPI<FormValues[FieldValueKey]>
{
    const { setFieldState, initialValues, submitted } = formApi;

    const inputReference = useRef<HTMLInputElement & HTMLFieldSetElement>(null);
    const [state, setState] = useState<FieldState<FormValues[FieldValueKey]>>({
        value: initialValues[name],
        touched: false,
        validState: 'NOT_VALIDATED',
        error: '',
        validity: null,
    })

    useEffect(() => {
        const inputElement = inputReference.current;

        // Clear custom validity from the last validation call.
        inputElement?.setCustomValidity?.('');

        const nativeAttributeValidation = inputElement?.checkValidity?.() ?? true;
        const validity = inputElement?.validity ?? null;

        // If we got error from native attr validation (required, pattern, type)
        // we still run validator in order to get some custom error message for
        // standard validation error if validator doesn't provide message we fallback
        // on standard validationMessage string [1] (ex. Please fill in input.)
        const nativeErrorMessage = inputElement?.validationMessage ?? ''
        const customValidation = validator(state.value, validity);

        if (customValidation || !nativeAttributeValidation) {
            // [1] Custom error message or fallback on native error message
            const validationMessage = customValidation || nativeErrorMessage;

            inputElement?.setCustomValidity?.(validationMessage)

            return setState(state => ({
                ...state,
                validState: 'INVALID',
                error: customValidation,
                validity
            }))
        }

        return setState(state => ({
            ...state,
            validState: 'VALID' as const,
            error: '',
            validity
        }))
    }, [state.value, validator]);

    // Sync field state with state on form level.
    useEffect(() => setFieldState(name, state), [name, state, setFieldState])

    return {
        input: {
            name: name.toString(),
            ref: inputReference,
            value: state.value,

            onBlur: () =>
                setState(state => ({...state, touched: true })),

            onChange: (event: ChangeEvent<HTMLInputElement> | FormValues[FieldValueKey]) => {
                const value = eventValue(event);

                setState(state => ({...state, value, }))
            },
        },
        meta: {
            ...state,
            touched: state.touched || submitted
        }
    }
}

function isChangeEvent<Value>(possibleEvent: ChangeEvent | Value): possibleEvent is ChangeEvent {
    return !!(possibleEvent as ChangeEvent).target;
}

function eventValue<Value>(event: ChangeEvent<HTMLInputElement> | Value ): Value {
    if (isChangeEvent(event)) {
        if (event.target.type === 'checkbox') {
            return event.target.checked as unknown as Value
        }

        return event.target.value as unknown as Value
    }

    return event;
}
