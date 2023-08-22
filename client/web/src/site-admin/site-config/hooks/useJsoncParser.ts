import { useState, useEffect, useCallback } from 'react'

import { ParseError, applyEdits, modify, parse } from 'jsonc-parser'

import { defaultModificationOptions } from '../SiteAdminConfigurationPage'

interface UseJsoncParserReturnType<T extends object> {
    /* parsed JSON */
    json?: T
    /* raw JSON */
    rawJson?: string
    /* error if parsing failed */
    error?: Error
    /* applies edits to JSON */
    update: (partOfJson: Partial<T>) => void
    /* resets to the original JSON state before any updates */
    reset: () => void
}

/**
 * React wrapper around 'jsonc-parser' to parse and modify settings JSON.
 * - It parses 'rawJsonValue' and stores local copies during updates as 'json' and 'rawJson'.
 * - if 'rawJsonValue' changes, it re-parses and updates 'json' and 'rawJson'.
 */
export function useJsoncParser<T extends object>(originalRawJson?: string): UseJsoncParserReturnType<T> {
    const [error, setError] = useState<Error>()
    const [json, setJson] = useState<T>()
    const [rawJson, setRawJson] = useState<string>()

    useEffect((): void => {
        setRawJson(originalRawJson)
    }, [originalRawJson])

    useEffect((): void => {
        if (!rawJson) {
            return
        }
        const errors: ParseError[] = []
        const parsedJson = parse(rawJson, errors, {
            allowTrailingComma: true,
            disallowComments: false,
        }) as T

        if (errors?.length > 0) {
            setError(new Error('Cannot parse JSON: ' + errors.join(', ')))
            setJson(undefined)
            return
        }
        setError(undefined)
        setJson(parsedJson)
    }, [rawJson])

    const update = useCallback((value: Partial<T>) => {
        setRawJson(prevRawJson => {
            if (!prevRawJson) {
                setError(new Error('No raw JSON to modify'))
                return
            }
            let newRawJson = prevRawJson
            for (const [fieldKey, fieldValue] of Object.entries(value)) {
                newRawJson = applyEdits(
                    newRawJson,
                    modify(newRawJson, [fieldKey], fieldValue, defaultModificationOptions)
                )
            }
            return newRawJson
        })
    }, [])

    const reset = useCallback(() => {
        setRawJson(originalRawJson)
    }, [originalRawJson])

    return { json, error, rawJson, update, reset }
}
