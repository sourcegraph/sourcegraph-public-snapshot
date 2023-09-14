import { type RefObject, useCallback, useRef } from 'react'

import { EMPTY, from, type Observable } from 'rxjs'
import { debounceTime, switchMap, tap } from 'rxjs/operators'

import { useEventObservable } from '../../../../../hooks'
import type { FieldMetaState } from '../useForm'
import { type AsyncValidator, getCustomValidationContext, getCustomValidationMessage } from '../validators'

const ASYNC_VALIDATION_DEBOUNCE_TIME = 500

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
    onValidationChange: (state: Partial<FieldMetaState<FieldValue>>) => void
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
     * Cancel handler to cancel ongoing async validation pipeline
     * */
    cancel: () => void

    /**
     * Start async validation pipeline.
     * */
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

    const asyncValidationPipeline = useCallback(
        (validationEvents: Observable<AsyncValidationEvent<FieldValue>>) =>
            validationEvents.pipe(
                debounceTime(ASYNC_VALIDATION_DEBOUNCE_TIME),
                tap(event => {
                    if (!event.canceled) {
                        const inputElement = inputReference.current

                        // Reset validation (native and custom) before async validation
                        inputElement?.setCustomValidity?.('')
                        onValidationChangeReference.current?.({
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
                tap(validationResult => {
                    const inputElement = inputReference.current
                    const validity = inputElement?.validity ?? null

                    if (validationResult) {
                        const validationMessage = getCustomValidationMessage(validationResult)
                        const validationContext = getCustomValidationContext(validationResult)

                        inputElement?.setCustomValidity?.(validationMessage)
                        onValidationChangeReference.current?.({
                            validState: 'INVALID',
                            error: validationMessage,
                            errorContext: validationContext,
                            validity,
                        })
                    } else {
                        onValidationChangeReference.current?.({
                            validState: 'VALID' as const,
                            error: '',
                            validity,
                        })
                    }
                })
            ),
        [inputReference, asyncValidator]
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
