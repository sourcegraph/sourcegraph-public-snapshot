import AJV from 'ajv'
import addFormats from 'ajv-formats'
import { load as loadYAML } from 'js-yaml'
import { debounce } from 'lodash'
import { useCallback, useEffect, useMemo, useState } from 'react'

import { useDebounce } from '@sourcegraph/wildcard'

import batchSpecSchemaJSON from '../../../../../../schema/batch_spec.schema.json'
import { BatchSpec } from '../../../schema/batch_spec.schema'

import { excludeRepo as excludeRepoFromYaml, hasOnOrImportChangesetsStatement } from './yaml-util'

const ajv = new AJV()
addFormats(ajv)

const formatError = (error: { instancePath: string; message?: string }): string => {
    if (!error.message) {
        return ''
    }

    // The error's instance path will have the format "/property1/property2", so we
    // convert each "/" to a "." and drop the first one.
    const dottedPath = error.instancePath.replace(/^\//, '').replace(/\//g, '.')
    return `${dottedPath} ${error.message}`
}

const DEBOUNCE_AMOUNT = 500

interface UseBatchSpecCodeResult {
    /** The current YAML code in the editor. */
    code: string
    /** The value of `code` but trail debounced by `DEBOUNCE_AMOUNT` */
    debouncedCode: string
    /** Callback to handle when the user modifies the code in the editor. */
    handleCodeChange: (newCode: string) => void
    /**
     * Boolean representing if `debouncedCode` is valid YAML code and satisfies the batch
     * spec schema requirements, or 'unknown' if validation has not yet recomputed.
     */
    isValid: boolean | 'unknown'
    /**
     * Any errors that occurred either while validating the batch spec YAML, or while
     * trying to automatically update it (i.e. to automatically exclude a repo).
     */
    errors: {
        validation?: string
        update?: string
    }
    /**
     * Method to automatically update the batch spec code with a modified `on: ` query to
     * exclude the provided `repo` at the provided `branch`
     *
     * NOTE: For monorepos, we currently will exclude all paths
     */
    excludeRepo: (repo: string, branch: string) => void
}

/**
 * Custom hook for "CreateOrEdit" page which packages up business logic and exposes an API
 * for managing the batch spec input YAML code that the user interacts with via the Monaco
 * editor.
 *
 * @param initialCode The initial YAML code that is displayed in the editor.
 * @param name The name of the batch change, which is used for validation.
 */
export const useBatchSpecCode = (initialCode: string, name: string): UseBatchSpecCodeResult => {
    const validateSpec = useMemo(() => {
        const schemaID = `${batchSpecSchemaJSON.$id}/${name}`

        const existingValidateFunction = ajv.getSchema(schemaID)
        if (existingValidateFunction) {
            return existingValidateFunction
        }

        // We enforce the exact name match on the schema. The user must use the settings
        // UI to change the name.
        const schemaJSONWithName = {
            ...batchSpecSchemaJSON,
            $id: schemaID,
            properties: {
                ...batchSpecSchemaJSON.properties,
                name: {
                    ...batchSpecSchemaJSON.properties.name,
                    pattern: `^${name}$`,
                },
            },
        }
        return ajv.compile<BatchSpec>(schemaJSONWithName)
    }, [name])

    const [code, setCode] = useState<string>(initialCode)
    const debouncedCode = useDebounce(code, 250)

    const [validationError, setValidationErrors] = useState<string>()
    const [updateError, setUpdateError] = useState<string>()

    const clearErrors = useCallback(() => {
        setValidationErrors(undefined)
        setUpdateError(undefined)
    }, [])

    const [isValid, setIsValid] = useState<boolean | 'unknown'>('unknown')

    const validate = useCallback(
        (newCode: string) => {
            try {
                const parsed = loadYAML(newCode)
                const valid = validateSpec(parsed)
                const hasOnOrImport = hasOnOrImportChangesetsStatement(newCode)
                setIsValid(valid && hasOnOrImport)
                if (!valid && validateSpec.errors?.length) {
                    setValidationErrors(
                        `The entered spec is invalid:\n  * ${validateSpec.errors.map(formatError).join('\n  * ')}`
                    )
                } else if (!hasOnOrImport) {
                    setValidationErrors(
                        'The entered spec must contain either an "on:" or "importingChangesets:" statement.'
                    )
                }
            } catch (error: unknown) {
                setIsValid(false)
                // Try to extract the error message.
                if (error && typeof error === 'object' && 'reason' in error) {
                    setValidationErrors((error as { reason: string }).reason)
                } else {
                    setValidationErrors('unknown validation error occurred')
                }
            }
        },
        [validateSpec]
    )

    // Run validation once for initial batch spec code.
    useEffect(() => validate(initialCode), [initialCode, validate])

    // Debounce validation to avoid excessive computation.
    const debouncedValidate = useMemo(() => debounce(validate, DEBOUNCE_AMOUNT), [validate])

    // Stop the debounced function on dismount.
    useEffect(
        () => () => {
            debouncedValidate.cancel()
        },
        [debouncedValidate]
    )

    const handleCodeChange = useCallback(
        (newCode: string) => {
            setCode(newCode)
            // We clear all errors and debounce validation on code change.
            clearErrors()
            setIsValid('unknown')
            debouncedValidate(newCode)
        },
        [debouncedValidate, clearErrors]
    )

    // Automatically updates the batch spec code when the user wants to exclude a repo
    // resolved in the workspaces preview.
    const excludeRepo = useCallback(
        (repo: string, branch: string) => {
            clearErrors()

            const result = excludeRepoFromYaml(code, repo, branch)

            if (result.success) {
                setCode(result.spec)
            } else {
                setUpdateError(
                    'Unable to update batch spec. Double-check to make sure there are no syntax errors, then try again.' +
                        result.error
                )
            }
        },
        [code, clearErrors]
    )

    return {
        code,
        debouncedCode,
        handleCodeChange,
        isValid,
        errors: {
            validation: validationError,
            update: updateError,
        },
        excludeRepo,
    }
}
