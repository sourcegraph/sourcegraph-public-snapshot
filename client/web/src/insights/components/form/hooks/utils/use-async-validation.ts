import { RefObject, useCallback, useRef } from 'react'
import { EMPTY, from, Observable } from 'rxjs'
import { debounceTime, switchMap, tap } from 'rxjs/operators'

import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'

import { FieldState, ValidationResult } from '../useForm'

const ASYNC_VALIDATION_DEBOUNCE_TIME = 500

export type AsyncValidator<FieldValue> = (
    value: FieldValue | undefined,
    validity: ValidityState | null
) => Promise<ValidationResult>

export interface UseAsyncValidationProps<FieldValue> {
    /**
     * Native input element used below to set validation state via html5
     * constraint validation API.
     * */
    inputReference: RefObject<HTMLInputElement>

    /**
     * Async validator, just an async function which returns or not returns
     * validation error.
     * */
    asyncValidator?: AsyncValidator<FieldValue>

    /**
     * Validation state change handler. Used below to update state of consumer
     * according to async logic aspects (mark field as touched, set validity state, etc).
     * */
    onValidationChange: (state: Partial<FieldState<FieldValue>>) => void
}

/**
 * Async validation event.
 * With that event consumer should start async validation pipeline.
 *
 * start(event: AsyncValidationEvent)
 * */
export interface AsyncValidationEvent<FieldValue> {
    /**
     * Value of form filed.
     */
    value: FieldValue | undefined

    /**
     * Constraint API validity state.
     */
    validity: ValidityState | null

    /**
     * Internal param to cancel all ongoing asyn validation processes.
     * used below to implement cancel handler from public API.
     * */
    canceled?: boolean
}

export interface useAsyncValidationAPI<FieldValue> {
    /**
     * cancel hander for cancalation ongoing
     * */
    cancel: () => void
    start: (event: AsyncValidationEvent<FieldValue>) => void
}

/**
 * Internal util hook for async validation. Part of implementation of useField hook.
 */
export function useAsyncValidation<FieldValue>(
    props: UseAsyncValidationProps<FieldValue>
): useAsyncValidationAPI<FieldValue> {
    const { inputReference, asyncValidator, onValidationChange } = props

    // Reference hack to allow consumers pass handlers without useMemo
    const onValidationChangeReference = useRef<UseAsyncValidationProps<FieldValue>['onValidationChange']>()
    onValidationChangeReference.current = onValidationChange

    const handleValidationChange = onValidationChangeReference.current

    const asyncValidationPipeline = useCallback(
        (validationEvents: Observable<AsyncValidationEvent<FieldValue>>) =>
            validationEvents.pipe(
                debounceTime(ASYNC_VALIDATION_DEBOUNCE_TIME),
                tap(event => {
                    if (!event.canceled) {
                        const inputElement = inputReference.current

                        // Reset validation (native and custom) before async validation
                        inputElement?.setCustomValidity?.('')
                        handleValidationChange({
                            validState: 'CHECKING',
                            touched: true,
                            error: '',
                            validity: inputElement?.validity ?? null,
                        })
                    }
                }),
                switchMap(({ value, validity, canceled }) =>
                    !canceled ? from(asyncValidator!(value, validity)) : EMPTY
                ),
                tap(validationMessage => {
                    const inputElement = inputReference.current
                    const validity = inputElement?.validity ?? null

                    if (validationMessage) {
                        console.log('Async validation result', validationMessage)

                        inputElement?.setCustomValidity?.(validationMessage)
                        handleValidationChange({
                            validState: 'INVALID',
                            error: validationMessage,
                            validity,
                        })
                    } else {
                        handleValidationChange({
                            validState: 'VALID' as const,
                            error: '',
                            validity,
                        })
                    }
                })
            ),
        [inputReference, handleValidationChange, asyncValidator]
    )

    const [startAsyncValidation] = useEventObservable(asyncValidationPipeline)

    const cancelAsyncValidation = useCallback(
        () =>
            startAsyncValidation({
                value: undefined,
                validity: null,
                canceled: true,
            }),
        [startAsyncValidation]
    )

    return {
        cancel: cancelAsyncValidation,
        start: startAsyncValidation,
    }
}
