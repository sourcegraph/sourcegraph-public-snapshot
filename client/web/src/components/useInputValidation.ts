import { compact, head } from 'lodash'
import { useMemo, useState, useCallback, useRef, useEffect } from 'react'
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
    synchronousValidators?: ((value: string) => string | undefined)[]

    /**
     * Optional array of asynchronous input validators. These must return
     * observables created with `fromFetch` for easy cancellation in `switchMap`.
     *
     * If there's no problem with the input, emit undefined. Else,
     * return with the reason the input is invalid.
     */
    asynchronousValidators?: ((value: string) => Observable<string | undefined>)[]
}

type ValidationResult = { kind: 'VALID' } | { kind: 'INVALID'; reason: string }

export type InputValidationState = { value: string; loading: boolean } & (
    | { kind: 'NOT_VALIDATED' }
    | { kind: 'VALID' }
    | { kind: 'INVALID'; reason: string }
)

/**
 * React hook to manage validation of a single form input field.
 * `useInputValidation` helps with coodinating the constraint validation API
 * and custom synchronous and asynchronous validators.
 *
 * @param options Config object that declares sync + async validators
 * @param initialValue
 * @param onInputChange Higher order function to execute side-effects given the latest input value and loading state.
 * Typically used to set state in a React component.
 * The function provided to `onInputChange` should be called with the previous input value and loading state
 *
 * @returns
 */
export function useInputValidation(
    options: ValidationOptions,
    onInputChange?: (inputStateCallback: (previousInputState: InputValidationState) => InputValidationState) => void
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

    const inputReference = useRef<HTMLInputElement | null>(null)

    const [inputState, setInputState] = useState({ value: options.initialValue ?? '', loading: false })

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
                    setInputState({ value, loading: asynchronousValidators.length > 0 })
                    onInputChange?.(() => ({ value, loading: asynchronousValidators.length > 0, kind: 'VALID' }))
                }),
                // Debounce everything.
                // This is to allow immediate validation on type but at the same time not flag invalid input as it's being typed.
                debounceTime(500),
                switchMap(value => {
                    // check validity (synchronous)
                    const valid = inputReference.current?.checkValidity()
                    if (!valid) {
                        return of({ kind: 'INVALID' as const, reason: inputReference.current?.validationMessage ?? '' })
                    }

                    // check any custom sync validators
                    const syncReason = head(compact(synchronousValidators.map(validator => validator(value))))
                    if (syncReason) {
                        inputReference.current?.setCustomValidity(syncReason)
                        return of({ kind: 'INVALID' as const, reason: syncReason })
                    }

                    if (asynchronousValidators.length === 0) {
                        // clear possible custom sync validation error from previous value
                        inputReference.current?.setCustomValidity('')
                        return of({ kind: 'VALID' as const })
                    }

                    // check async validators
                    return zip(...asynchronousValidators.map(validator => validator(value))).pipe(
                        map(values => head(compact(values))),
                        map(reason => (reason ? { kind: 'INVALID' as const, reason } : { kind: 'VALID' as const })),
                        tap(result => {
                            if (result.kind === 'INVALID') {
                                inputReference.current?.setCustomValidity(result.reason)
                            } else {
                                inputReference.current?.setCustomValidity('')
                            }
                            onInputChange?.(previousInputState => ({ ...previousInputState, loading: false }))
                            setInputState(previousInputState => ({ ...previousInputState, loading: false }))
                        })
                    )
                }),
                tap(() => {
                    onInputChange?.(previousInputState => ({ ...previousInputState, loading: false }))
                    setInputState(previousInputState => ({ ...previousInputState, loading: false }))
                }),
                catchError(error => of({ kind: 'INVALID' as const, reason: asError(error).message || 'Unknown error' }))
            ),
        [synchronousValidators, asynchronousValidators, onInputChange, options.initialValue]
    )

    const [nextInputChangeEvent, validationResult] = useEventObservable(validationPipeline)

    return [
        validationResult ? { ...inputState, ...validationResult } : { ...inputState, kind: 'NOT_VALIDATED' },
        nextInputChangeEvent,
        inputReference,
    ]
}

/**
 * Derives className based on validation state. Use with `useInputValidation`.
 *
 * @param inputState
 */
export function deriveInputClassName(inputState: InputValidationState): string {
    if (inputState.loading || inputState.kind === 'NOT_VALIDATED') {
        return ''
    }
    return inputState.kind === 'INVALID' ? 'is-invalid' : 'is-valid'
}
