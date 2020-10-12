import { compact, head } from 'lodash'
import { useMemo, useState, useCallback, useRef } from 'react'
import { concat, EMPTY, Observable, of, zip } from 'rxjs'
import { catchError, map, switchMap, tap, debounceTime } from 'rxjs/operators'
import { useEventObservable } from './useObservable'
import { asError } from './errors'

/**
 * Configuration used by `useInputValidation`
 */
export interface ValidationOptions {
    initialValue?: string

    /**
     * Optional array of synchronous input validators.
     *
     * If there's no problem with the input, return undefined. Else,
     * return with the reason the input is invalid.
     */
    synchronousValidators?: ((value: string) => ValidationResult)[]

    /**
     * Optional array of asynchronous input validators. These must return
     * observables created with `fromFetch` for easy cancellation in `switchMap`.
     *
     * If there's no problem with the input, emit undefined. Else,
     * return with the reason the input is invalid.
     */
    asynchronousValidators?: ((value: string) => Observable<ValidationResult>)[]
}

type ValidationResult = string | undefined

export type InputValidationState = { value: string } & (
    | { kind: 'NOT_VALIDATED' | 'LOADING' | 'VALID' }
    | { kind: 'INVALID'; reason: string }
)

/**
 * React hook to manage validation of a single form input field.
 * `useInputValidation` helps with coodinating the constraint validation API
 * and custom synchronous and asynchronous validators.
 *
 * @param options Config object that declares sync + async validators
 * @param initialValue
 *
 * @returns
 */
export function useInputValidation(
    options: ValidationOptions
): [
    InputValidationState,
    (change: React.ChangeEvent<HTMLInputElement>) => void,
    React.MutableRefObject<HTMLInputElement | null>
] {
    const { synchronousValidators, asynchronousValidators } = useMemo(() => {
        const { synchronousValidators = [], asynchronousValidators = [] } = options

        return {
            synchronousValidators,
            asynchronousValidators,
        }
    }, [options])

    const inputReference = useRef<HTMLInputElement>(null)

    const [inputState, setInputState] = useState<InputValidationState>({
        kind: 'NOT_VALIDATED',
        value: options.initialValue ?? '',
    })

    const validationPipeline = useCallback(
        (events: Observable<React.ChangeEvent<HTMLInputElement>>): Observable<ValidationResult> =>
            concat(
                options.initialValue !== undefined ? of(options.initialValue) : EMPTY,
                events.pipe(
                    tap(event => event.preventDefault()),
                    map(event => event.target.value)
                )
            ).pipe(
                tap(value => {
                    inputReference.current?.setCustomValidity('')
                    setInputState({ value, kind: 'LOADING' })
                }),
                // Debounce everything.
                // This is to allow immediate validation on type but at the same time not flag invalid input as it's being typed.
                debounceTime(500),
                switchMap(value => {
                    // check validity (synchronous)
                    const valid = inputReference.current?.checkValidity()
                    if (!valid) {
                        setInputState({
                            value,
                            kind: 'INVALID',
                            reason: inputReference.current?.validationMessage ?? '',
                        })
                        return of(inputReference.current?.validationMessage ?? '')
                    }

                    // check custom sync validators
                    const syncReason = head(compact(synchronousValidators.map(validator => validator(value))))
                    if (syncReason) {
                        inputReference.current?.setCustomValidity(syncReason)

                        setInputState({
                            value,
                            kind: 'INVALID',
                            reason: syncReason,
                        })
                        return of(syncReason)
                    }

                    if (asynchronousValidators.length === 0) {
                        // clear possible custom sync validation error from previous value
                        inputReference.current?.setCustomValidity('')
                        setInputState({
                            value,
                            kind: 'VALID',
                        })
                        return of(undefined)
                    }

                    // check async validators
                    return zip(...asynchronousValidators.map(validator => validator(value))).pipe(
                        map(values => head(compact(values))),
                        tap(reason => {
                            inputReference.current?.setCustomValidity(reason ?? '')
                            setInputState(reason ? { kind: 'INVALID', value, reason } : { kind: 'VALID', value })
                        })
                    )
                }),
                catchError(error => of(asError(error).message || 'Unknown error'))
            ),
        [synchronousValidators, asynchronousValidators, options.initialValue]
    )

    const [nextInputChangeEvent] = useEventObservable(validationPipeline)

    return [inputState, nextInputChangeEvent, inputReference]
}

/**
 * Derives className based on validation state. Use with `useInputValidation`.
 *
 * @param inputState
 */
export function deriveInputClassName(inputState: InputValidationState): string {
    if (inputState.kind === 'LOADING' || inputState.kind === 'NOT_VALIDATED') {
        return ''
    }
    return inputState.kind === 'INVALID' ? 'is-invalid' : 'is-valid'
}
