import { Validator } from './hooks/useField'
import { ValidationResult } from './hooks/useForm'

/**
 * Validator for required form field which returns error massage
 * as a sign of invalid state.
 * */
export const createRequiredValidator = <Value>(errorMessage: string): Validator<Value> => (value, validity) => {
    if (validity?.valueMissing) {
        return errorMessage
    }

    // Handle the string value case.
    if (typeof value === 'string' && value.trim() === '') {
        return errorMessage
    }

    return
}

/**
 * Composes a few validators together and show first error for form field.
 * */
export const composeValidators = <Value>(...validators: Validator<Value>[]): Validator<Value> => (value, validity) =>
    validators.reduce<ValidationResult>((error, validator) => error || validator(value, validity), undefined)
