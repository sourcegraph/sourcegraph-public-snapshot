import React, { useCallback, useMemo, useState } from 'react'

import { noop } from 'lodash'

import { EditBatchChangeFields } from '../../../graphql-operations'
import { useBatchSpecCode, UseBatchSpecCodeResult } from '../create/useBatchSpecCode'

export interface BatchSpecContextErrors {
    // Errors from trying to automatically apply updates to the batch spec code.
    codeUpdate?: string | Error
    // Errors from validating the batch spec code.
    codeValidation?: string | Error
    // Errors from trying to preview the workspaces that would be affected by the batch spec.
    preview?: string | Error
    // Errors from trying to execute the batch spec.
    execute?: string | Error
}

export interface BatchSpecContextState {
    readonly batchChange: EditBatchChangeFields
    readonly batchSpec: EditBatchChangeFields['currentSpec'] & { isApplied: boolean }

    readonly editor: UseBatchSpecCodeResult

    readonly errors: BatchSpecContextErrors
    setError: (type: keyof BatchSpecContextErrors, error: string | Error | undefined) => void
}

export const defaultState = (): BatchSpecContextState => ({
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    batchChange: {} as EditBatchChangeFields,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    batchSpec: {} as EditBatchChangeFields['currentSpec'] & { isApplied: boolean },
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    editor: {} as UseBatchSpecCodeResult,
    errors: {},
    setError: noop,
})

/**
 * TODO:
 *
 * @see BatchSpecContextProvider
 */
export const BatchSpecContext = React.createContext<BatchSpecContextState>(defaultState())

interface BatchSpecContextProviderProps {
    batchChange: EditBatchChangeFields
}

export const BatchSpecContextProvider: React.FunctionComponent<BatchSpecContextProviderProps> = ({
    children,
    batchChange,
}) => {
    const {
        currentSpec,
        batchSpecs: { nodes },
    } = batchChange

    // The first node from the batch specs is the latest batch spec for a batch change. If
    // it's different from the `currentSpec` on the batch change, that means the latest
    // batch spec has not yet been applied.
    const latestSpec = nodes[0] || currentSpec
    // TODO: This should probably just be a field on GraphQL.
    const isLatestSpecApplied = useMemo(() => currentSpec.id === latestSpec.id, [currentSpec.id, latestSpec.id])

    const editor = useBatchSpecCode(latestSpec.originalInput, batchChange.name)

    // TODO: This will probably go away
    const [errors, setErrors] = useState<Pick<BatchSpecContextErrors, 'preview' | 'execute'>>({})
    const setError = useCallback((type: keyof BatchSpecContextErrors, error: string | Error | undefined) => {
        setErrors(errors => ({ ...errors, [type]: error }))
    }, [])

    const { handleCodeChange } = editor
    const clearPreviewErrorsAndHandleCodeChange = useCallback(
        (newCode: string) => {
            setError('preview', undefined)
            handleCodeChange(newCode)
        },
        [handleCodeChange, setError]
    )

    return (
        <BatchSpecContext.Provider
            value={{
                batchChange,
                batchSpec: { ...latestSpec, isApplied: isLatestSpecApplied },
                editor: { ...editor, handleCodeChange: clearPreviewErrorsAndHandleCodeChange },
                errors: { ...errors, codeUpdate: editor.errors.update, codeValidation: editor.errors.validation },
                setError,
            }}
        >
            {children}
        </BatchSpecContext.Provider>
    )
}
