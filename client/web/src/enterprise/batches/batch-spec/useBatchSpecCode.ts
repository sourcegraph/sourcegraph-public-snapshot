import { useCallback, useEffect, useMemo, useState } from 'react'

import AJV from 'ajv'
import addFormats from 'ajv-formats'
import { load as loadYAML } from 'js-yaml'
import { debounce } from 'lodash'

import type { BatchSpec } from '@sourcegraph/shared/src/schema/batch_spec.schema'
import { useDebounce } from '@sourcegraph/wildcard'

import batchSpecSchemaJSON from '../../../../../../schema/batch_spec.schema.json'

import helloWorldSample from './edit/library/hello-world.batch.yaml'
import {
    excludeRepo as excludeRepoFromYaml,
    hasOnOrImportChangesetsStatement,
    isMinimalBatchSpec,
    insertNameIntoLibraryItem,
} from './yaml-util'

const ajv = new AJV()
addFormats(ajv)

const formatError = (error: { instancePath: string; message?: string }): string => {
    if (!error.message) {
        return ''
    }

    // The error's instance path will have the format "/property1/property2", so we
    // convert each "/" to a "." and drop the first one.
    const dottedPath = error.instancePath.replace(/^\//, '').replaceAll('/', '.')
    return `${dottedPath} ${error.message}`
}

const DEBOUNCE_AMOUNT = 500

/** API for state managing the batch spec input YAML code in the Monaco editor. */
export interface UseBatchSpecCodeResult {
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
     * Whether or not the batch spec YAML on the server which was used to preview
     * workspaces is up-to-date with that which is presently in the editor.
     */
    isServerStale: boolean
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
 * Custom hook for edit page which packages up business logic and exposes an API for
 * managing the batch spec input YAML code that the user interacts with via the Monaco
 * editor.
 *
 * @param originalInput The initial YAML code of the batch spec.
 * @param name The name of the batch change, which is used for validation.
 */
export const useBatchSpecCode = (originalInput: string, name: string): UseBatchSpecCodeResult => {
    const validateFunction = useMemo(() => {
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

    const validate = useCallback(
        (code: string): [isValid: boolean, error?: string] => {
            try {
                const parsed = loadYAML(code)
                const valid = validateFunction(parsed)
                const hasOnOrImport = hasOnOrImportChangesetsStatement(code)

                const validationError =
                    !valid && validateFunction.errors?.length
                        ? `The entered spec is invalid:\n  * ${validateFunction.errors.map(formatError).join('\n  * ')}`
                        : !hasOnOrImport
                        ? 'The entered spec must contain either an "on:" or "importingChangesets:" statement.'
                        : undefined

                return [valid && hasOnOrImport, validationError]
            } catch (error: unknown) {
                // Try to extract the error message.
                const validationError =
                    error && typeof error === 'object' && 'reason' in error
                        ? (error as { reason: string }).reason
                        : 'Unknown validation error occurred.'

                return [false, validationError]
            }
        },
        [validateFunction]
    )

    const [code, setCode] = useState<string>(() =>
        // Start with the hello world sample code initially if the user hasn't written any
        // batch spec code yet, otherwise show the latest spec code.
        isMinimalBatchSpec(originalInput) ? insertNameIntoLibraryItem(helloWorldSample, name) : originalInput
    )
    const debouncedCode = useDebounce(code, DEBOUNCE_AMOUNT)

    const [validationError, setValidationErrors] = useState<string | undefined>(() => validate(code)[1])
    const [updateError, setUpdateError] = useState<string | undefined>()

    const clearErrors = useCallback(() => {
        setValidationErrors(undefined)
        setUpdateError(undefined)
    }, [])

    const [isValid, setIsValid] = useState<boolean | 'unknown'>(() => validate(code)[0])

    const revalidate = useCallback(
        (newCode: string) => {
            const [isValid, validationError] = validate(newCode)
            setIsValid(isValid)
            setValidationErrors(validationError)
        },
        [validate]
    )

    // Debounce revalidation to avoid excessive computation.
    const debouncedValidate = useMemo(() => debounce(revalidate, DEBOUNCE_AMOUNT), [revalidate])

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
    // TODO: https://github.com/sourcegraph/sourcegraph/issues/25085
    const excludeRepo = useCallback(
        (repo: string, branch: string) => {
            clearErrors()

            const result = excludeRepoFromYaml(code, repo, branch)

            if (result.success) {
                setCode(result.spec)
            } else {
                setUpdateError(
                    'Unable to update batch spec. Double-check to make sure there are no syntax errors, then try again. ' +
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
        // NOTE: The batch spec YAML code is considered stale if any part of it changes.
        // This is because of a current limitation of the backend where we need to
        // re-submit the batch spec code and wait for the new workspaces preview to finish
        // resolving before we can execute, or else the execution will use an older batch
        // spec. We will address this when we implement the "auto-saving" feature and
        // decouple previewing workspaces from updating the batch spec code.
        isServerStale: originalInput !== debouncedCode,
        errors: {
            validation: validationError,
            update: updateError,
        },
        excludeRepo,
    }
}
