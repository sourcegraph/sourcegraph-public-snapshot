
export type ValidationResult = string | undefined;
export type Validator = (value: string) => ValidationResult;

export const createRequiredValidator = (errorMessage: string) =>
    (value: string): ValidationResult => (value ? undefined : errorMessage);

export const createValidRegExpValidator = (errorMessage: string) =>
    (value: string): ValidationResult => {
        try {
            new RegExp(value);

            return;
        } catch {
            return errorMessage;
        }
    }

export const composeValidators = (...validators: Validator[]) =>
    (value: string): ValidationResult =>
        validators.reduce<ValidationResult>(
            (error, validator) => error || validator(value),
            undefined
        );
