export type ValidationResult = string | undefined
export type Validator = (value: string) => ValidationResult

/** Validator for form field which returns error massage as a sign of invalid state. */
export const createRequiredValidator = (errorMessage: string) => (value: string): ValidationResult =>
    value ? undefined : errorMessage

/** Special validator to check field with regexp as a value of input. */
export const createValidRegExpValidator = (errorMessage: string) => (value: string): ValidationResult => {
    try {
        new RegExp(value)

        return
    } catch {
        return errorMessage
    }
}

/** Composes a few validators together and show first error for form field. */
export const composeValidators = (...validators: Validator[]) => (value: string): ValidationResult =>
    validators.reduce<ValidationResult>((error, validator) => error || validator(value), undefined)
