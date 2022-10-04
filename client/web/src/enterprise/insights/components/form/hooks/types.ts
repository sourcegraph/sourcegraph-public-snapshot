
type ValidationError = Error | string

interface SuccessValidationResult {
    isValid: true
    errors: never
}

interface FailedValidationResult {
    isValid: false
    errors: ValidationError[]
}

export type ValidationResult = SuccessValidationResult | FailedValidationResult

export interface ValidationInput<FieldValue> {
    /** Field value, it has no value (undefined) if input hasn't been touched. */
    value: FieldValue | undefined,

    /**
     * Native input validation state, if input has native browser validation attributes such as
     * required, minLength, maxLength, regexp validation this state will provide it.
     */
    validity: ValidityState | null
}

export type Validator<FieldValue> = (input: ValidationInput<FieldValue>) => ValidationResult
export type AsyncValidator<FieldValue> = (input: ValidationInput<FieldValue>) => Promise<ValidationResult>
