import AJV from 'ajv'
import addFormats from 'ajv-formats'
import classNames from 'classnames'
import { load as loadYAML } from 'js-yaml'
import { debounce } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory } from 'react-router'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { BatchSpecWorkspaceResolutionState, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useMutation } from '@sourcegraph/shared/src/graphql/graphql'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { ButtonTooltip } from '@sourcegraph/web/src/components/ButtonTooltip'
import { Container, PageHeader, useDebounce } from '@sourcegraph/wildcard'

import batchSpecSchemaJSON from '../../../../../../schema/batch_spec.schema.json'
import { BatchChangesIcon } from '../../../batches/icons'
import {
    CreateBatchSpecFromRawResult,
    CreateBatchSpecFromRawVariables,
    ReplaceBatchSpecInputResult,
    ReplaceBatchSpecInputVariables,
} from '../../../graphql-operations'
import { BatchSpec } from '../../../schema/batch_spec.schema'
import { Settings } from '../../../schema/settings.schema'
import { BatchSpecDownloadLink } from '../BatchSpec'

import { CREATE_BATCH_SPEC_FROM_RAW, executeBatchSpec, fetchBatchSpec, REPLACE_BATCH_SPEC_INPUT } from './backend'
import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import helloWorldSample from './examples/hello-world.batch.yaml'
import { NamespaceSelector } from './NamespaceSelector'
import styles from './NewCreateBatchChangePage.module.scss'
import { useNamespaces } from './useNamespaces'
import { WorkspacesPreview } from './WorkspacesPreview'
import { excludeRepo as excludeRepoFromYaml } from './yaml-util'

const ajv = new AJV()
addFormats(ajv)
const VALIDATE_SPEC = ajv.compile<BatchSpec>(batchSpecSchemaJSON)

const getNamespaceDisplayName = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return namespace.displayName ?? namespace.username
        case 'Org':
            return namespace.displayName ?? namespace.name
    }
}

/** TODO: This duplicates the URL field from the org/user resolvers on the backend, but we
 * don't have access to that from the settings cascade presently. Can we get it included
 * in the cascade instead somehow? */
const getNamespaceBatchChangesURL = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return '/users/' + namespace.username + '/batch-changes'
        case 'Org':
            return '/organizations/' + namespace.name + '/batch-changes'
    }
}

interface CreateBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {}

export const NewCreateBatchChangePage: React.FunctionComponent<CreateBatchChangePageProps> = ({
    isLightTheme,
    settingsCascade,
}) => {
    const { namespaces, defaultSelectedNamespace } = useNamespaces(settingsCascade)

    // The namespace selected for creating the new batch spec under.
    const [selectedNamespace, setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
        defaultSelectedNamespace
    )

    const { code, debouncedCode, isValid, handleCodeChange, excludeRepo, errors } = useBatchSpecCode(helloWorldSample)

    const [serverError, setServerError] = useState<Error>()

    const clearErrorsAndHandleCodeChange = useCallback(
        (newCode: string) => {
            setServerError(undefined)
            handleCodeChange(newCode)
        },
        [handleCodeChange]
    )

    const [
        createBatchSpecFromRaw,
        { data: createBatchSpecFromRawData, loading: createBatchSpecFromRawLoading },
    ] = useMutation<CreateBatchSpecFromRawResult, CreateBatchSpecFromRawVariables>(CREATE_BATCH_SPEC_FROM_RAW, {})

    const [
        replaceBatchSpecInput,
        { data: replaceBatchSpecInputData, loading: replaceBatchSpecInputLoading },
    ] = useMutation<ReplaceBatchSpecInputResult, ReplaceBatchSpecInputVariables>(REPLACE_BATCH_SPEC_INPUT, {})

    const currentBatchSpecID = useMemo(
        () =>
            // If we have replaced the batch spec input already, the initial batch spec
            // has been superseded, so prefer that one.
            replaceBatchSpecInputData?.replaceBatchSpecInput.id ||
            createBatchSpecFromRawData?.createBatchSpecFromRaw.id,
        [replaceBatchSpecInputData, createBatchSpecFromRawData]
    )

    const preview = useCallback(() => {
        setServerError(undefined)

        // If we have a batch spec ID already, we're replacing the exiting batch spec with
        // a new one.
        if (currentBatchSpecID) {
            return replaceBatchSpecInput({ variables: { spec: code, previousSpec: currentBatchSpecID } }).catch(
                setServerError
            )
        }

        // Otherwise, we're creating a new batch spec from the raw spec input YAML.
        return createBatchSpecFromRaw({
            variables: { spec: code, namespace: selectedNamespace.id },
        }).catch(setServerError)
    }, [code, currentBatchSpecID, selectedNamespace, createBatchSpecFromRaw, replaceBatchSpecInput])

    // const history = useHistory()
    // const submitBatchSpec = useCallback<React.MouseEventHandler>(async () => {
    //     if (!previewID) {
    //         return
    //     }
    //     setIsLoading(true)
    //     try {
    //         const execution = await executeBatchSpec(previewID)
    //         history.push(`${execution.namespace.url}/batch-changes/executions/${execution.id}`)
    //     } catch (error) {
    //         setIsLoading(error)
    //     }
    // }, [previewID, history])

    const [canExecute, executionTooltip] = useMemo(() => {
        const canExecute = isValid && !serverError && !createBatchSpecFromRawLoading && !replaceBatchSpecInputLoading
        const executionTooltip =
            !isValid || serverError
                ? "There's a problem with your batch spec."
                : createBatchSpecFromRawLoading || replaceBatchSpecInputLoading
                ? 'Wait for the preview to finish.'
                : undefined

        return [canExecute, executionTooltip]
    }, [isValid, serverError, createBatchSpecFromRawLoading, replaceBatchSpecInputLoading])

    return (
        <div className="d-flex flex-column p-4 w-100 h-100">
            <div className="d-flex flex-0 justify-content-between">
                <div className="flex-1">
                    <PageHeader
                        path={[
                            { icon: BatchChangesIcon },
                            {
                                to: getNamespaceBatchChangesURL(selectedNamespace),
                                text: getNamespaceDisplayName(selectedNamespace),
                            },
                            { text: 'Create batch change' },
                        ]}
                        className="flex-1 pb-2"
                        description="Run custom code over hundreds of repositories and manage the resulting changesets."
                    />

                    <NamespaceSelector
                        namespaces={namespaces}
                        selectedNamespace={selectedNamespace.id}
                        onSelect={setSelectedNamespace}
                    />
                </div>
                <div className="d-flex flex-column flex-0 align-items-center justify-content-center">
                    <ButtonTooltip
                        type="button"
                        className="btn btn-primary mb-2"
                        // onClick={submitBatchSpec}
                        disabled={!canExecute}
                        tooltip={executionTooltip}
                    >
                        Run batch spec
                    </ButtonTooltip>
                    <BatchSpecDownloadLink name="new-batch-spec" originalInput={code}>
                        or download for src-cli
                    </BatchSpecDownloadLink>
                </div>
            </div>
            <div className="d-flex flex-1">
                <div className={styles.editorContainer}>
                    <MonacoBatchSpecEditor
                        isLightTheme={isLightTheme}
                        value={code}
                        onChange={clearErrorsAndHandleCodeChange}
                    />
                </div>
                <Container className={styles.workspacesPreviewContainer}>
                    {errors.update && <ErrorAlert error={errors.update} />}
                    {!isValid && errors.validation.length > 0 && (
                        <ErrorAlert error={`The entered spec is invalid:\n  * ${errors.validation.join('\n  * ')}`} />
                    )}
                    {serverError && <ErrorAlert error={serverError} />}
                    <WorkspacesPreview batchSpecInput={debouncedCode} disabled={isValid !== true} preview={preview} />
                </Container>
            </div>
        </div>
    )
}

interface UseBatchSpecCodeResult {
    code: string
    debouncedCode: string
    handleCodeChange: (newCode: string) => void
    isValid: boolean | 'unknown'
    errors: {
        validation: string[]
        update?: string
    }
    excludeRepo: (repo: string, branch: string) => void
}

const useBatchSpecCode = (initialCode: string): UseBatchSpecCodeResult => {
    const [code, setCode] = useState<string>(initialCode)
    const debouncedCode = useDebounce(code, 250)

    const [validationErrors, setValidationErrors] = useState<string[]>([])
    const [updateError, setUpdateError] = useState<string>()

    const clearErrors = useCallback(() => {
        setValidationErrors([])
        setUpdateError(undefined)
    }, [])

    const [isValid, setIsValid] = useState<boolean | 'unknown'>('unknown')

    const validate = useCallback((newCode: string) => {
        try {
            const parsed = loadYAML(newCode)
            const valid = VALIDATE_SPEC(parsed)
            setIsValid(valid)
            if (!valid && VALIDATE_SPEC.errors) {
                setValidationErrors(VALIDATE_SPEC.errors.map(error => error.message || '') || [])
            }
        } catch (error: unknown) {
            setIsValid(false)
            if (error && typeof error === 'object' && 'reason' in error) {
                setValidationErrors([(error as { reason: string }).reason])
            } else {
                setValidationErrors(['unknown validation error occurred'])
            }
        }
    }, [])

    // Run validation once for initial batch spec code
    useEffect(() => validate(initialCode), [initialCode, validate])

    // Debounce validation to avoid excessive computation.
    const debouncedValidate = useMemo(() => debounce(validate, 250), [validate])

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
            clearErrors()
            setIsValid('unknown')
            debouncedValidate(newCode)
        },
        [debouncedValidate, clearErrors]
    )

    // Updates the batch spec code when the user wants to exclude a repo resolved in the
    // workspaces preview.
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
            validation: validationErrors,
            update: updateError,
        },
        excludeRepo,
    }
}
