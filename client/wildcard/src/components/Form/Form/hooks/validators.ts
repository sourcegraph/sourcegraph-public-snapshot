/**
 * If validator returns nothing, that means that validation has passed successfully.
 */
export type SuccessValidationResult<Context> = undefined | void | SuccessContextValidationResult<Context>

/**
 * Context-based success validation result, context is used when validation has passed
 * successfully, but you also want to show some success validation details, like what
 * validation exactly has passed.
 */
interface SuccessContextValidationResult<Context> {
    context?: Context
}

/**
 * Failed validation result supports two formats, if validator returns
 * - string (legacy format) this will be used as validation error
 * - enhanced failed result contains error message and failed validation context (generic useful information)
 */
export type FailedValidationResult<Context> =
    | string
    | {
          errorMessage: string
          context?: Context
      }

export type ValidationResult<Context = unknown> = SuccessValidationResult<Context> | FailedValidationResult<Context>

export type Validator<FieldValue, Context = unknown> = (
    value: FieldValue | undefined,
    validity: ValidityState | null
) => ValidationResult<Context>

export type AsyncValidator<FieldValue> = (
    value: FieldValue | undefined,
    validity: ValidityState | null
) => Promise<ValidationResult<unknown>>

/**
 * Helper function, extracts validation message from different validation formats
 */
export function getCustomValidationMessage<Context>(validationResult: ValidationResult<Context>): string {
    if (!validationResult) {
        return ''
    }

    // Legacy validation result format
    if (typeof validationResult === 'string') {
        return validationResult
    }

    if ('errorMessage' in validationResult) {
        return validationResult.errorMessage
    }

    return ''
}

/**
 * Helper function, extracts validation error context from different validation formats
 */
export function getCustomValidationContext<Context>(validationResult: ValidationResult<Context>): Context | undefined {
    if (!validationResult) {
        return
    }

    // Legacy validation result format
    if (typeof validationResult === 'string') {
        return
    }

    if ('context' in validationResult) {
        return validationResult.context
    }

    return
}

/**
 * Compose sync validators in single validator, this validator stops validation as soon as it gets
 * first error from validators sequence.
 *
 * @param validators - validators sequence
 */
export function composeValidators<FieldValue, Context>(
    validators: Validator<FieldValue, Context>[]
): Validator<FieldValue, Context> {
    return (value: FieldValue | undefined, validity: ValidityState | null): ValidationResult<Context> => {
        for (const validator of validators) {
            const validationResult = validator(value, validity)

            if (!validationResult) {
                continue
            }

            return validationResult
        }

        // Successful validation
        return
    }
}

/**
 * Enhanced field validator, that enhances native field validation (required attribute)
 * and handles trim string case in case of string-like field value.
 */
export function createRequiredValidator<FieldValue, Context>(errorMessage: string): Validator<FieldValue, Context> {
    return (value: FieldValue | undefined, validity: ValidityState | null): ValidationResult<Context> => {
        if (validity?.valueMissing || (typeof value === 'string' && value.trim() === '')) {
            return errorMessage
        }

        // Successful validation
        return
    }
}
