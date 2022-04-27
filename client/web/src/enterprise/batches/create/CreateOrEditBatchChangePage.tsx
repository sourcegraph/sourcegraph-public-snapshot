import React, { useCallback, useMemo, useState } from 'react'

import { ApolloQueryResult } from '@apollo/client'
import classNames from 'classnames'
import { noop } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import LockIcon from 'mdi-react/LockIcon'
import { useHistory } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    PageHeader,
    Button,
    Container,
    Input,
    LoadingSpinner,
    FeedbackBadge,
    Icon,
    Panel,
    Tabs,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    RadioButton,
} from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
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
import { BatchSpecDownloadLink } from '../BatchSpec'

import { GET_BATCH_CHANGE_TO_EDIT, CREATE_EMPTY_BATCH_CHANGE } from './backend'
import { DownloadSpecModal } from './DownloadSpecModal'
import { EditorFeedbackPanel } from './editor/EditorFeedbackPanel'
import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import { ExecutionOptions, ExecutionOptionsDropdown } from './ExecutionOptions'
import { LibraryPane } from './library/LibraryPane'
import { NamespaceSelector } from './NamespaceSelector'
import { useBatchSpecCode } from './useBatchSpecCode'
import { useExecuteBatchSpec } from './useExecuteBatchSpec'
import { useInitialBatchSpec } from './useInitialBatchSpec'
import { useNamespaces } from './useNamespaces'
import { useWorkspacesPreview } from './useWorkspacesPreview'
import { useImportingChangesets } from './workspaces-preview/useImportingChangesets'
import { useWorkspaces, WorkspacePreviewFilters } from './workspaces-preview/useWorkspaces'
import { WorkspacesPreview } from './workspaces-preview/WorkspacesPreview'

import styles from './CreateOrEditBatchChangePage.module.scss'

export interface CreateOrEditBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {
    /**
     * The id for the namespace that the batch change should be created in, or that it
     * already belongs to, if it already exists.
     */
    initialNamespaceID?: Scalars['ID']
    /** The batch change name, if it already exists. */
    batchChangeName?: BatchChangeFields['name']
    /** Display the configuration page in read-only mode */
    isReadOnly?: boolean
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
        return <CreatePage isReadOnly={false} namespaceID={initialNamespaceID} {...props} />
    }

    if (loading && !data) {
        return (
            <div className="w-100 text-center">
                <Icon className="m-2" as={LoadingSpinner} />
            </div>
        )
    }

    if (!data?.batchChange || error) {
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    if (props.isReadOnly) {
        return (
            <CreatePage
                isReadOnly={props.isReadOnly}
                batchChangeName={data?.batchChange.name}
                namespaceID={data?.batchChange.namespace.id}
                {...props}
            />
        )
    }

    return <EditPage batchChange={data.batchChange} refetchBatchChange={refetchBatchChange} {...props} />
}

interface CreatePageProps extends SettingsCascadeProps<Settings> {
    /**
     * The namespace the batch change should be created in. If none is provided, it will
     * default to the user's own namespace.
     */
    namespaceID?: Scalars['ID']
    /** Display configuration in read-only mode */
    isReadOnly?: boolean
    /** Name of batch change when in read-only mode */
    batchChangeName?: string
}

const CreatePage: React.FunctionComponent<CreatePageProps> = props => {
    const isNewBatchChange = props.batchChangeName === undefined && !props.isReadOnly

    return (
        <div className="w-100 p-4">
            <PageTitle title="Create new batch change" />
            <PageHeader
                path={[{ icon: BatchChangesIcon, to: '.' }, { text: 'Create batch change' }]}
                className="flex-1 pb-2"
                description="Run custom code over hundreds of repositories and manage the resulting changesets."
                annotation={<FeedbackBadge status="experimental" feedback={{ mailto: 'support@sourcegraph.com' }} />}
            />
            <Tabs>
                <TabList>
                    <Tab className="text-content py-2 px-3">1. Configuration</Tab>
                    <Tab className="text-content py-2 px-3" disabled={isNewBatchChange}>
                        2. Batch spec
                    </Tab>
                    <Tab className="text-content py-2 px-3" disabled={isNewBatchChange}>
                        3. Execution
                    </Tab>
                    <Tab className="text-content py-2 px-3" disabled={isNewBatchChange}>
                        4. Preview
                    </Tab>
                </TabList>
                <TabPanels>
                    <TabPanel>
                        <BatchConfigurationPage {...props} />
                    </TabPanel>

                    <TabPanel>
                        <div>Batch spec</div>
                    </TabPanel>

                    <TabPanel>
                        <div>Execution</div>
                    </TabPanel>

                    <TabPanel>
                        <div>Preview</div>
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </div>
    )
}

interface BatchConfigurationPageProps extends SettingsCascadeProps<Settings> {
    /**
     * The namespace the batch change should be created in. If none is provided, it will
     * default to the user's own namespace.
     */
    namespaceID?: Scalars['ID']
    /** Display configuration in read-only mode */
    isReadOnly?: boolean
    /** Batch change when in read-only mode */
    batchChangeName?: string
}

const BatchConfigurationPage: React.FunctionComponent<BatchConfigurationPageProps> = ({
    namespaceID,
    settingsCascade,
    isReadOnly,
    batchChangeName,
}) => {
    const [createEmptyBatchChange, { loading, error }] = useMutation<
        CreateEmptyBatchChangeResult,
        CreateEmptyBatchChangeVariables
    >(CREATE_EMPTY_BATCH_CHANGE)

    const { namespaces, defaultSelectedNamespace } = useNamespaces(settingsCascade, namespaceID)

    // The namespace selected for creating the new batch change under.
    const [selectedNamespace, setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
        defaultSelectedNamespace
    )

    const [nameInput, setNameInput] = useState(batchChangeName || '')
    const [isNameValid, setIsNameValid] = useState<boolean>()

    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setNameInput(event.target.value)
        setIsNameValid(NAME_PATTERN.test(event.target.value))
    }, [])

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
        <Form className={styles.batchConfigForm} onSubmit={handleCreate}>
            <Container className="mb-4">
                {error && <ErrorAlert error={error} />}
                <NamespaceSelector
                    namespaces={namespaces}
                    selectedNamespace={selectedNamespace.id}
                    onSelect={setSelectedNamespace}
                    disabled={isReadOnly}
                />
                <Input
                    label="Batch change name"
                    value={nameInput}
                    onChange={onNameChange}
                    pattern={String(NAME_PATTERN)}
                    required={true}
                    status={isNameValid === undefined ? undefined : isNameValid ? 'valid' : 'error'}
                    placeholder="My batch change name"
                    disabled={isReadOnly}
                />
                <small className="text-muted">
                    Give it a short, descriptive name to reference the batch change on Sourcegraph. Do not include
                    confidential information.{' '}
                    <span className={classNames(isNameValid === false && 'text-danger')}>
                        Only regular characters, _ and - are allowed.
                    </span>
                </small>
                <hr className="my-3" />
                <h3 className="text-muted">
                    Visibility <Icon data-tooltip="Coming soon" as={InfoCircleOutlineIcon} />
                </h3>
                <div className="form-group mb-1">
                    <RadioButton
                        name="visibility"
                        value="public"
                        className="mr-2"
                        checked={true}
                        disabled={true}
                        label="Public"
                        aria-label="Public"
                    />
                </div>
                <div className="form-group mb-0">
                    <RadioButton
                        name="visibility"
                        value="private"
                        className="mr-2 mb-0"
                        disabled={true}
                        label={
                            <>
                                Private <Icon className="text-warning" aria-hidden={true} as={LockIcon} />
                            </>
                        }
                        aria-label="Private"
                    />
                </div>
            </Container>

            {!isReadOnly && (
                <div className={styles.ctaGroup}>
                    <Button variant="secondary" type="button" outline={true} onClick={handleCancel}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        type="submit"
                        onClick={handleCreate}
                        disabled={loading || nameInput === '' || !isNameValid}
                    >
                        Create
                    </Button>
                </div>
            )}
        </Form>
    )
}

const INVALID_BATCH_SPEC_TOOLTIP = "There's a problem with your batch spec."
const WORKSPACES_PREVIEW_SIZE = 'batch-changes.ssbc-workspaces-preview-size'

interface EditPageProps extends ThemeProps {
    batchChange: EditBatchChangeFields
    refetchBatchChange: () => Promise<ApolloQueryResult<GetBatchChangeToEditResult>>
}

const EditPage: React.FunctionComponent<EditPageProps> = ({ batchChange, refetchBatchChange, isLightTheme }) => {
    // Get the latest batch spec for the batch change.
    const { batchSpec, isApplied: isLatestBatchSpecApplied, initialCode: initialBatchSpecCode } = useInitialBatchSpec(
        batchChange
    )

    // Manage the batch spec input YAML code that's being edited.
    const { code, debouncedCode, isValid, handleCodeChange, excludeRepo, errors: codeErrors } = useBatchSpecCode(
        initialBatchSpecCode,
        batchChange.name
    )

    const [filters, setFilters] = useState<WorkspacePreviewFilters>()
    const [isDownloadSpecModalOpen, setIsDownloadSpecModalOpen] = useState(false)
    const [downloadSpecModalDismissed, setDownloadSpecModalDismissed] = useTemporarySetting(
        'batches.downloadSpecModalDismissed',
        false
    )

    const workspacesConnection = useWorkspaces(batchSpec.id, filters)
    const importingChangesetsConnection = useImportingChangesets(batchSpec.id)

    // When we successfully submit the latest batch spec code to the backend for a new
    // workspaces preview, we follow up by refetching the batch change to get the latest
    // batch spec ID.
    const onComplete = useCallback(() => {
        // We handle any error here higher up the chain, so we can ignore it.
        refetchBatchChange().then(noop).catch(noop)
    }, [refetchBatchChange])

    // NOTE: Technically there's only one option, and it's actually a preview option.
    const [executionOptions, setExecutionOptions] = useState<ExecutionOptions>({ runWithoutCache: false })

    // Manage the batch spec that was last submitted to the backend for the workspaces preview.
    const {
        preview: previewBatchSpec,
        isInProgress: isWorkspacesPreviewInProgress,
        error: previewError,
        clearError: clearPreviewError,
        hasPreviewed,
        cancel,
        resolutionState,
    } = useWorkspacesPreview(batchSpec.id, {
        isBatchSpecApplied: isLatestBatchSpecApplied,
        namespaceID: batchChange.namespace.id,
        noCache: executionOptions.runWithoutCache,
        onComplete,
        filters,
    })

    const clearErrorsAndHandleCodeChange = useCallback(
        (newCode: string) => {
            clearPreviewError()
            handleCodeChange(newCode)
        },
        [handleCodeChange, clearPreviewError]
    )

    // Disable the preview button if the batch spec code is invalid or the on: statement
    // is missing, or if we're already processing a preview.
    const previewDisabled = useMemo(
        () => (isValid !== true ? INVALID_BATCH_SPEC_TOOLTIP : isWorkspacesPreviewInProgress),
        [isValid, isWorkspacesPreviewInProgress]
    )

    // The batch spec YAML code is considered stale if any part of it changes. This is
    // because of a current limitation of the backend where we need to re-submit the batch
    // spec code and wait for the new workspaces preview to finish resolving before we can
    // execute, or else the execution will use an older batch spec. We will address this
    // when we implement the "auto-saving" feature and decouple previewing workspaces from
    // updating the batch spec code.
    const isBatchSpecStale = useMemo(() => initialBatchSpecCode !== debouncedCode, [
        initialBatchSpecCode,
        debouncedCode,
    ])

    // Manage submitting a batch spec for execution.
    const { executeBatchSpec, isLoading: isExecuting, error: executeError } = useExecuteBatchSpec(batchSpec.id)

    // Disable the execute button if any of the following are true:
    // - The batch spec code is invalid.
    // - There was an error with the preview.
    // - We're in the middle of previewing or executing the batch spec.
    // - We haven't yet submitted the batch spec to the backend yet for a preview.
    // - The batch spec on the backend is stale.
    // - The current workspaces evaluation is not complete.
    const [isExecutionDisabled, executionTooltip] = useMemo(() => {
        const isExecutionDisabled = Boolean(
            isValid !== true ||
                previewError ||
                isWorkspacesPreviewInProgress ||
                isExecuting ||
                !hasPreviewed ||
                isBatchSpecStale ||
                resolutionState !== BatchSpecWorkspaceResolutionState.COMPLETED
        )
        // The execution tooltip only shows if the execute button is disabled, and explains why.
        const executionTooltip =
            isValid === false || previewError
                ? INVALID_BATCH_SPEC_TOOLTIP
                : !hasPreviewed
                ? 'Preview workspaces first before you run.'
                : isBatchSpecStale
                ? 'Update your workspaces preview before you run.'
                : isWorkspacesPreviewInProgress || resolutionState !== BatchSpecWorkspaceResolutionState.COMPLETED
                ? 'Wait for the preview to finish first.'
                : undefined

        return [isExecutionDisabled, executionTooltip]
    }, [
        hasPreviewed,
        isValid,
        previewError,
        isWorkspacesPreviewInProgress,
        isExecuting,
        isBatchSpecStale,
        resolutionState,
    ])

    const actionButtons = (
        <>
            <ExecutionOptionsDropdown
                execute={executeBatchSpec}
                isExecutionDisabled={isExecutionDisabled}
                executionTooltip={executionTooltip}
                options={executionOptions}
                onChangeOptions={setExecutionOptions}
            />

            {downloadSpecModalDismissed ? (
                <BatchSpecDownloadLink name={batchChange.name} originalInput={code} isLightTheme={isLightTheme}>
                    or download for src-cli
                </BatchSpecDownloadLink>
            ) : (
                <Button className={styles.downloadLink} variant="link" onClick={() => setIsDownloadSpecModalOpen(true)}>
                    or download for src-cli
                </Button>
            )}
        </>
    )

    return (
        <BatchChangePage
            namespace={batchChange.namespace}
            title={batchChange.name}
            description={batchChange.description}
            actionButtons={actionButtons}
        >
            <div className={classNames(styles.editorLayoutContainer, 'd-flex flex-1 mt-2')}>
                <LibraryPane name={batchChange.name} onReplaceItem={clearErrorsAndHandleCodeChange} />
                <div className={styles.editorContainer}>
                    <h4 className={styles.header}>Batch spec</h4>
                    <MonacoBatchSpecEditor
                        batchChangeName={batchChange.name}
                        className={styles.editor}
                        isLightTheme={isLightTheme}
                        value={code}
                        onChange={clearErrorsAndHandleCodeChange}
                    />
                    <EditorFeedbackPanel
                        errors={{
                            codeUpdate: codeErrors.update,
                            codeValidation: codeErrors.validation,
                            preview: previewError,
                            execute: executeError,
                        }}
                    />

                    {isDownloadSpecModalOpen && !downloadSpecModalDismissed ? (
                        <DownloadSpecModal
                            name={batchChange.name}
                            originalInput={code}
                            isLightTheme={isLightTheme}
                            setDownloadSpecModalDismissed={setDownloadSpecModalDismissed}
                            setIsDownloadSpecModalOpen={setIsDownloadSpecModalOpen}
                        />
                    ) : null}
                </div>
                <Panel
                    className="d-flex"
                    defaultSize={500}
                    minSize={405}
                    maxSize={1400}
                    position="right"
                    storageKey={WORKSPACES_PREVIEW_SIZE}
                >
                    <div className={styles.workspacesPreviewContainer}>
                        <WorkspacesPreview
                            previewDisabled={previewDisabled}
                            preview={() => previewBatchSpec(debouncedCode)}
                            batchSpecStale={
                                isBatchSpecStale || isWorkspacesPreviewInProgress || resolutionState === 'CANCELED'
                            }
                            hasPreviewed={hasPreviewed}
                            excludeRepo={excludeRepo}
                            cancel={cancel}
                            isWorkspacesPreviewInProgress={isWorkspacesPreviewInProgress}
                            resolutionState={resolutionState}
                            workspacesConnection={workspacesConnection}
                            importingChangesetsConnection={importingChangesetsConnection}
                            setFilters={setFilters}
                        />
                    </div>
                </Panel>
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
            {actionButtons && (
                <div className="d-flex flex-column flex-0 align-items-center justify-content-center">
                    {actionButtons}
                </div>
            )}
        </div>
        {children}
    </div>
)
/* Regex pattern for a valid batch change name. Needs to match what's defined in the BatchSpec JSON schema. */
const NAME_PATTERN = /^[\w.-]+$/
