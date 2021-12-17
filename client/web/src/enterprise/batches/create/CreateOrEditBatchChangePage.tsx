import { ApolloQueryResult } from '@apollo/client'
import classNames from 'classnames'
import { noop } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { useHistory } from 'react-router'

import { useMutation, useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { ButtonTooltip } from '@sourcegraph/web/src/components/ButtonTooltip'
import { HeroPage } from '@sourcegraph/web/src/components/HeroPage'
import { PageHeader, Button, Container, Input, LoadingSpinner } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { FeedbackBadge } from '../../../components/FeedbackBadge'
import {
    BatchChangeFields,
    EditBatchChangeFields,
    GetBatchChangeToEditResult,
    GetBatchChangeToEditVariables,
    CreateEmptyBatchChangeVariables,
    CreateEmptyBatchChangeResult,
    Scalars,
    BatchSpecWorkspaceResolutionState,
} from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'
import { BatchSpecDownloadLink } from '../BatchSpec'

import { GET_BATCH_CHANGE_TO_EDIT, CREATE_EMPTY_BATCH_CHANGE } from './backend'
import styles from './CreateOrEditBatchChangePage.module.scss'
import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import { LibraryPane } from './library/LibraryPane'
import { NamespaceSelector } from './NamespaceSelector'
import { useBatchSpecCode } from './useBatchSpecCode'
import { usePreviewBatchSpec } from './useBatchSpecPreview'
import { useExecuteBatchSpec } from './useExecuteBatchSpec'
import { useInitialBatchSpec } from './useInitialBatchSpec'
import { useNamespaces } from './useNamespaces'
import { useBatchSpecWorkspaceResolution, WorkspacesPreview } from './workspaces-preview/WorkspacesPreview'

export interface CreateOrEditBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {
    /**
     * The id for the namespace that the batch change should be created in, or that it
     * already belongs to, if it already exists.
     */
    initialNamespaceID?: Scalars['ID']
    /** The batch change name, if it already exists. */
    batchChangeName?: BatchChangeFields['name']
}

/**
 * CreateOrEditBatchChangePage is the new SSBC-oriented page for creating a new batch change
 * or editing and re-executing a new batch spec for an existing one.
 */
export const CreateOrEditBatchChangePage: React.FunctionComponent<CreateOrEditBatchChangePageProps> = ({
    initialNamespaceID,
    batchChangeName,
    ...props
}) => {
    const { data, error, loading, refetch } = useQuery<GetBatchChangeToEditResult, GetBatchChangeToEditVariables>(
        GET_BATCH_CHANGE_TO_EDIT,
        {
            // If we don't have the batch change name or namespace, the user hasn't created a
            // batch change yet, so skip the request.
            skip: !initialNamespaceID || !batchChangeName,
            variables: {
                namespace: initialNamespaceID as Scalars['ID'],
                name: batchChangeName as BatchChangeFields['name'],
            },
            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',
        }
    )

    const refetchBatchChange = useCallback(
        () =>
            refetch({
                namespace: initialNamespaceID as Scalars['ID'],
                name: batchChangeName as BatchChangeFields['name'],
            }),
        [initialNamespaceID, batchChangeName, refetch]
    )

    if (!batchChangeName) {
        return <CreatePage namespaceID={initialNamespaceID} {...props} />
    }

    if (loading && !data) {
        return (
            <div className="w-100 text-center">
                <LoadingSpinner className="icon-inline m-2" />
            </div>
        )
    }

    if (!data?.batchChange || error) {
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    return <EditPage batchChange={data.batchChange} refetchBatchChange={refetchBatchChange} {...props} />
}

interface CreatePageProps extends SettingsCascadeProps<Settings> {
    /**
     * The namespace the batch change should be created in. If none is provided, it will
     * default to the user's own namespace.
     */
    namespaceID?: Scalars['ID']
}

const CreatePage: React.FunctionComponent<CreatePageProps> = ({ namespaceID, settingsCascade }) => {
    const [createEmptyBatchChange, { loading, error }] = useMutation<
        CreateEmptyBatchChangeResult,
        CreateEmptyBatchChangeVariables
    >(CREATE_EMPTY_BATCH_CHANGE)

    const { namespaces, defaultSelectedNamespace } = useNamespaces(settingsCascade, namespaceID)

    // The namespace selected for creating the new batch change under.
    const [selectedNamespace, setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
        defaultSelectedNamespace
    )

    const [nameInput, setNameInput] = useState('')

    const history = useHistory()
    const handleCancel = (): void => history.goBack()
    const handleCreate = (): void => {
        createEmptyBatchChange({
            variables: { namespace: selectedNamespace.id, name: nameInput },
        })
            .then(({ data }) => (data ? history.push(`${data.createEmptyBatchChange.url}/edit`) : noop()))
            // We destructure and surface the error from `useMutation` instead.
            .catch(noop)
    }

    return (
        <BatchChangePage namespace={selectedNamespace} title="Create batch change">
            <div className={styles.settingsContainer}>
                <h4>Batch spec settings</h4>
                <Container>
                    {error && <ErrorAlert error={error} />}
                    <NamespaceSelector
                        namespaces={namespaces}
                        selectedNamespace={selectedNamespace.id}
                        onSelect={setSelectedNamespace}
                    />
                    <Input
                        className={styles.nameInput}
                        label="Batch change name"
                        value={nameInput}
                        onChange={event => setNameInput(event.target.value)}
                        onKeyPress={event => {
                            if (event.key === 'Enter') {
                                handleCreate()
                            }
                        }}
                    />
                </Container>
                <div className="mt-3 align-self-end">
                    <Button variant="secondary" outline={true} className="mr-2" onClick={handleCancel}>
                        Cancel
                    </Button>
                    <Button variant="primary" onClick={handleCreate} disabled={loading}>
                        Create
                    </Button>
                </div>
            </div>
        </BatchChangePage>
    )
}

interface EditPageProps extends ThemeProps, SettingsCascadeProps<Settings> {
    batchChange: EditBatchChangeFields
    refetchBatchChange: () => Promise<ApolloQueryResult<GetBatchChangeToEditResult>>
}

const EditPage: React.FunctionComponent<EditPageProps> = ({
    batchChange,
    refetchBatchChange,
    isLightTheme,
    settingsCascade,
}) => {
    // Get the latest batch spec for the batch change.
    const { batchSpec, isApplied: isLatestBatchSpecApplied, initialCode: initialBatchSpecCode } = useInitialBatchSpec(
        batchChange
    )

    // TODO: Only needed when edit name/namespace form is open
    // Get the namespaces this user has access to.
    const { namespaces: _namespaces, defaultSelectedNamespace } = useNamespaces(
        settingsCascade,
        batchChange.namespace.id
    )

    // TODO: Only needed when edit name/namespace form is open
    // The namespace selected for creating the new batch spec under.
    const [selectedNamespace, _setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
        defaultSelectedNamespace
    )

    const [noCache, setNoCache] = useState<boolean>(false)

    const onChangeNoCache = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            setNoCache(event.target.checked)
            // Mark that the batch spec code on the backend is now stale.
            setBatchSpecStale(true)
        },
        [setNoCache]
    )

    // Manage the batch spec input YAML code that's being edited.
    const { code, debouncedCode, isValid, handleCodeChange, excludeRepo, errors: codeErrors } = useBatchSpecCode(
        initialBatchSpecCode
    )

    // Track whenever the batch spec code that is presently in the editor is newer than
    // the batch spec code on the backend.
    const [batchSpecStale, setBatchSpecStale] = useState(false)

    // When we successfully submit the latest batch spec code to the backend for a new
    // workspaces preview, we can:
    // - Mark that the batch spec code is no longer stale.
    // - Mark that the user has previewed the workspaces at least once.
    // - Refetch the batch change to get the latest batch spec.
    const onCompletePreview = useCallback(() => {
        setBatchSpecStale(false)
        refetchBatchChange().then(noop).catch(noop)
    }, [refetchBatchChange])

    // Manage the batch spec that was last submitted to the backend for the workspaces preview.
    const {
        previewBatchSpec,
        isLoading: isLoadingPreview,
        error: previewError,
        clearError: clearPreviewError,
        hasPreviewed,
    } = usePreviewBatchSpec(
        batchSpec.id,
        isLatestBatchSpecApplied,
        batchChange.namespace.id,
        noCache,
        onCompletePreview
    )

    const clearErrorsAndHandleCodeChange = useCallback(
        (newCode: string) => {
            clearPreviewError()
            // Mark that the batch spec code on the backend is now stale.
            setBatchSpecStale(true)
            handleCodeChange(newCode)
        },
        [handleCodeChange, clearPreviewError]
    )

    // Disable the preview button if the batch spec code is invalid or the on: statement
    // is missing, or if we're already processing a preview.
    const previewDisabled = useMemo(() => isValid !== true || isLoadingPreview, [isValid, isLoadingPreview])

    const workspacesPreviewResolution = useBatchSpecWorkspaceResolution(batchSpec, { fetchPolicy: 'cache-first' })

    // Manage submitting a batch spec for execution.
    const { executeBatchSpec, isLoading: isExecuting, error: executeError } = useExecuteBatchSpec(batchSpec.id)

    // Disable the execute button if any of the following are true:
    // - The batch spec code is invalid.
    // - There was an error with the preview.
    // - We're in the middle of previewing or executing the batch spec.
    // - We haven't yet submitted the batch spec to the backend yet for a preview.
    // - The batch spec on the backend is stale.
    // - The current workspaces evaluation is not complete.
    const [disableExecution, executionTooltip] = useMemo(() => {
        const disableExecution = Boolean(
            isValid !== true ||
                previewError ||
                isLoadingPreview ||
                isExecuting ||
                !hasPreviewed ||
                batchSpecStale ||
                workspacesPreviewResolution?.state !== BatchSpecWorkspaceResolutionState.COMPLETED
        )
        // The execution tooltip only shows if the execute button is disabled, and explains why.
        const executionTooltip =
            isValid === false || previewError
                ? "There's a problem with your batch spec."
                : !hasPreviewed
                ? 'Preview workspaces first before you run.'
                : batchSpecStale
                ? 'Update your workspaces preview before you run.'
                : isLoadingPreview || workspacesPreviewResolution?.state !== BatchSpecWorkspaceResolutionState.COMPLETED
                ? 'Wait for the preview to finish first.'
                : undefined

        return [disableExecution, executionTooltip]
    }, [
        hasPreviewed,
        isValid,
        previewError,
        isLoadingPreview,
        isExecuting,
        batchSpecStale,
        workspacesPreviewResolution?.state,
    ])

    const errors =
        codeErrors.update || codeErrors.validation || previewError || executeError ? (
            <div className="w-100">
                {codeErrors.update && <ErrorAlert error={codeErrors.update} />}
                {codeErrors.validation && <ErrorAlert error={codeErrors.validation} />}
                {previewError && <ErrorAlert error={previewError} />}
                {executeError && <ErrorAlert error={executeError} />}
            </div>
        ) : null

    const buttons = (
        <>
            <ButtonTooltip
                type="button"
                className="btn btn-primary mb-2"
                onClick={executeBatchSpec}
                disabled={disableExecution}
                tooltip={executionTooltip}
            >
                Run batch spec
            </ButtonTooltip>
            <BatchSpecDownloadLink name="new-batch-spec" originalInput={code} isLightTheme={isLightTheme}>
                or download for src-cli
            </BatchSpecDownloadLink>
            <div className="form-group">
                <label>
                    <input type="checkbox" className="mr-2" checked={noCache} onChange={onChangeNoCache} />
                    Disable cache
                </label>
            </div>
        </>
    )

    return (
        <BatchChangePage
            namespace={selectedNamespace}
            title={batchChange.name}
            description={batchChange.description}
            actionButtons={buttons}
        >
            <div className={classNames(styles.editorLayoutContainer, 'd-flex flex-1')}>
                <LibraryPane name={batchChange.name} onReplaceItem={clearErrorsAndHandleCodeChange} />
                <div className={styles.editorContainer}>
                    <h4>Batch spec</h4>
                    <MonacoBatchSpecEditor
                        className={styles.editor}
                        isLightTheme={isLightTheme}
                        value={code}
                        onChange={clearErrorsAndHandleCodeChange}
                    />
                </div>
                <div
                    className={classNames(
                        styles.workspacesPreviewContainer,
                        'd-flex flex-column align-items-center pl-4'
                    )}
                >
                    {errors}
                    <WorkspacesPreview
                        batchSpec={batchSpec}
                        hasPreviewed={hasPreviewed}
                        previewDisabled={previewDisabled}
                        preview={() => previewBatchSpec(debouncedCode)}
                        batchSpecStale={batchSpecStale}
                        excludeRepo={excludeRepo}
                    />
                </div>
            </div>
        </BatchChangePage>
    )
}

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

interface BatchChangePageProps {
    /** The namespace that should appear in the topmost `PageHeader`. */
    namespace: SettingsUserSubject | SettingsOrgSubject
    /** The title to use in the topmost `PageHeader`, alongside the `namespaceName`. */
    title: string
    /** The description to use in the topmost `PageHeader` beneath the titles. */
    description?: string | null
    /** Optionally, any action buttons that should appear in the top left of the page. */
    actionButtons?: JSX.Element
}

/**
 * BatchChangePage is a page layout component that renders a consistent header for
 * SSBC-style batch change pages and should wrap the other content contained on the page.
 */
const BatchChangePage: React.FunctionComponent<BatchChangePageProps> = ({
    children,
    namespace,
    title,
    description,
    actionButtons,
}) => (
    <div className="d-flex flex-column p-4 w-100 h-100">
        <div className="d-flex flex-0 justify-content-between align-items-start">
            <PageHeader
                path={[
                    { icon: BatchChangesIcon },
                    {
                        to: getNamespaceBatchChangesURL(namespace),
                        text: getNamespaceDisplayName(namespace),
                    },
                    { text: title },
                ]}
                className="flex-1 pb-2"
                description={
                    description || 'Run custom code over hundreds of repositories and manage the resulting changesets.'
                }
                annotation={<FeedbackBadge status="experimental" feedback={{ mailto: 'support@sourcegraph.com' }} />}
            />
            <div className="d-flex flex-column flex-0 align-items-center justify-content-center">{actionButtons}</div>
        </div>
        {children}
    </div>
)
