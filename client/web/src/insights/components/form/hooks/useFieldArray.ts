import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { noop } from 'rxjs';

import { useFieldAPI } from './useField';
import { FieldErrorState, FormAPI, Validator } from './useForm';
import { getEventValue } from './utils/get-event-value';
import { AsyncValidator } from './utils/use-async-validation';

export interface Validators<FieldValue> {
    sync?: Validator<FieldValue>,
    async?: AsyncValidator<FieldValue>
}

export type FieldItem<ComplexField> = { [P in keyof ComplexField] : useFieldAPI<ComplexField[P]> }

export interface UseFieldArrayAPI<ComplexField> {
    items: FieldItem<ComplexField>[]
    values: ComplexField[]
    meta: FieldErrorState
    add: () => void
}

export function useFieldArray<
    FormValues,
    FieldKey extends keyof FormAPI<FormValues>['initialValues'],
    FieldValue = FormValues[FieldKey] extends (infer R)[] ? R : never>(
        name: FormValues[FieldKey] extends unknown[] ? FieldKey : never,
        formAPI: FormAPI<FormValues>,
        defaultValues: FieldValue,
        config: { [P in keyof FieldValue]?: Validator<FieldValue[P]>},
        validator: Validator<FormValues[FieldKey]>): UseFieldArrayAPI<FieldValue> {

    const {
        setFieldState: setFormFieldState,
        initialValues: formInitialValues,
        submitted: formSubmitted,
        touched: formTouched } = formAPI

    // Use useRef for form api handler in order to avoid unnecessary
    // calls if API handler has been changed.
    const setFieldStateReference = useRef<FormAPI<FormValues>['setFieldState']>(setFormFieldState)
    setFieldStateReference.current = setFormFieldState

    const fieldReferences = useRef<Record<keyof FieldValue, HTMLInputElement | null>[]>([])
    const initialValues = formInitialValues[name] as unknown as FieldValue[]

    const [state, setState] = useState<Pick<FieldErrorState, 'validState' | 'error'>>({
        validState: 'NOT_VALIDATED',
        error: null
    })

    const [itemsValues, setItemsValues] = useState<Record<keyof FieldValue, any>[]>(
        () => initialValues.map(initialValue => {
            const fieldKeys = Object.keys(initialValue) as (keyof FieldValue)[];

            return fieldKeys.reduce((fields, filedKey) => {
                fields[filedKey] = initialValue[filedKey]

                return fields
            }, {} as Record<keyof FieldValue, any>)
        })
    )

    const [itemsState, setItemsState] = useState<Record<keyof FieldValue, FieldErrorState>[]>(
        () => initialValues.map(initialValue => {
            const fieldKeys = Object.keys(initialValue) as (keyof FieldValue)[];

            return fieldKeys.reduce((fields, filedKey) => {
                fields[filedKey] = {
                    touched: false,
                    validState: 'NOT_VALIDATED',
                    error: null,
                    validity: null,
                }

                return fields
            }, {} as Record<keyof FieldValue, FieldErrorState>)
        }))

    // Update particular field of particular item at fieldArray state
    const setFieldState = useCallback((
        index: number,
        name: keyof FieldValue,
        newFieldState: Partial<FieldErrorState>
    ) => {
        setItemsState(items => {
            const itemState = items[index];
            const fieldState = itemState[name]

            return [
                ...items.slice(0, index),
                { ...itemState, [name]: { ...fieldState, ...newFieldState } },
                ...items.slice(index + 1),
            ]
        })
    }, [])

    const items = useMemo<FieldItem<FieldValue>[]>(
        () => itemsState.map((itemState, index) => {
            const fieldKeys = Object.keys(itemState) as (keyof FieldValue)[];
            const itemValue = itemsValues[index];

            return fieldKeys.reduce((fields, key) => {
                fields[key] = {
                    input: {
                        ref: input => {
                            if (!fieldReferences.current[index]) {
                                fieldReferences.current[index] = {} as Record<keyof FieldValue, HTMLInputElement | null>
                            }

                            fieldReferences.current[index][key] = input
                        },
                        name: '',
                        value: itemValue[key].value,
                        onBlur: () => setFieldState(index, key, { touched: true }),
                        onChange: event => {
                            const value = getEventValue(event);

                            setItemsValues(itemsValues => {
                                const item = itemsValues[index];

                                return [
                                    ...itemsValues.slice(0, index),
                                    { ...item, [key]: value },
                                    ...itemsValues.slice(index + 1),
                                ]
                            })
                        }
                    },
                    meta: itemState[key]
                }

                return fields
            }, {} as FieldItem<FieldValue>)
        }),
        [itemsState, setFieldState, itemsValues]
    )

    useEffect(() => {
        for (let index = 0; index < itemsValues.length; index++) {
            const item = itemsValues[index];
            const fieldKeys = Object.keys(item) as (keyof FieldValue)[]

            for (const fieldKey of fieldKeys) {
                const inputElement = fieldReferences.current[index]?.[fieldKey]

                inputElement?.setCustomValidity?.('')

                const nativeAttributeValidation = inputElement?.checkValidity?.() ?? true
                const validity = inputElement?.validity ?? null
                const nativeErrorMessage = inputElement?.validationMessage ?? ''
                const validator = config[fieldKey] ?? noop
                const customValidation = validator(item[fieldKey], validity)

                if (customValidation || !nativeAttributeValidation) {
                    const validationMessage = customValidation || nativeErrorMessage

                    inputElement?.setCustomValidity?.(validationMessage)

                    setFieldState(index, fieldKey, {
                        validState: 'INVALID',
                        error: validationMessage,
                        validity,
                    })
                } else {
                    setFieldState(index, fieldKey, {
                        validState: 'VALID' as const,
                        error: '',
                        validity,
                    })
                }
            }
        }

        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [...itemsValues, config])

    useEffect(() => {
        const validationResult = validator(itemsValues as unknown as FormValues[FieldKey], null)

        // Self validation
        if (validationResult) {
            return setState({
                error: validationResult,
                validState: 'INVALID'
            })
        }

        return setState({
            error: null,
            validState: 'VALID'
        })

        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [...itemsValues, validator, setState])

    // Sync field state with state on form level - useForm hook will used this state to run
    // onSubmit handler and track validation state to prevent onSubmit run when async
    // validation is going.
    useEffect(
        () => {
            setFieldStateReference.current(name, {
                value: itemsValues,
                touched: false,
                validState: state.validState,
                error: state.error,
                validity: null
            })
        },
        [name, itemsValues, state]
    )

    const addItem = useCallback(() => {
        const fieldKeys = Object.keys(defaultValues) as (keyof FieldValue)[];

        const fieldsState = fieldKeys.reduce((fields, filedKey) => {
            fields[filedKey] = {
                touched: false,
                validState: 'NOT_VALIDATED',
                error: null,
                validity: null,
            }

            return fields
        }, {} as Record<keyof FieldValue, FieldErrorState>)

        const itemValue = fieldKeys.reduce((fields, filedKey) => {
            fields[filedKey] = defaultValues[filedKey]

            return fields
        }, {} as Record<keyof FieldValue, any>)

        setItemsState(items => [...items, fieldsState])
        setItemsValues(itemsValues => [...itemsValues, itemValue])
    }, [defaultValues])

    return {
        values: itemsValues,
        items,
        meta: {
            touched: formSubmitted || formTouched,
            validState: state.validState,
            error: state.error,
            validity: null
        },
        add: addItem
    }
}
