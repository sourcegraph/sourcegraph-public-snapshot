import { useCallback, useEffect, useState, type FormEventHandler, type FunctionComponent } from 'react'

import { mdiLink } from '@mdi/js'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import type { TelemetryRecorder, TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Alert,
    Button,
    ErrorAlert,
    Form,
    H3,
    Icon,
    Link,
    LoadingSpinner,
    Modal,
    PageHeader,
    screenReaderAnnounce,
} from '@sourcegraph/wildcard'

import type {
    TransferWorkflowOwnershipResult,
    TransferWorkflowOwnershipVariables,
    UpdateWorkflowResult,
    UpdateWorkflowVariables,
    WorkflowFields,
    WorkflowResult,
    WorkflowVariables,
} from '../graphql-operations'
import { NamespaceSelector } from '../namespaces/NamespaceSelector'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'

import { WorkflowForm, type WorkflowFormValue } from './Form'
import {
    deleteWorkflowMutation,
    transferWorkflowOwnershipMutation,
    updateWorkflowMutation,
    workflowQuery,
} from './graphql'
import { WorkflowPage } from './Page'

/**
 * Page to edit a workflow.
 */
export const EditPage: FunctionComponent<TelemetryV2Props & { isSourcegraphDotCom: boolean }> = ({
    telemetryRecorder,
    isSourcegraphDotCom,
}) => {
    const { id } = useParams<{ id: string }>()
    const { data, loading, error } = useQuery<WorkflowResult, WorkflowVariables>(workflowQuery, {
        variables: { id: id! },
    })
    const workflow = data?.node?.__typename === 'Workflow' ? data.node : null

    return (
        <WorkflowPage
            title={workflow ? `Editing: ${workflow.description} - workflow` : 'Edit workflow'}
            actions={
                workflow && (
                    <Button to={workflow.url} variant="secondary" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiLink} /> Permalink
                    </Button>
                )
            }
            breadcrumbsNamespace={workflow?.owner}
            breadcrumbs={<PageHeader.Breadcrumb>Edit</PageHeader.Breadcrumb>}
        >
            {loading ? (
                <LoadingSpinner />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : !workflow ? (
                <Alert variant="danger" as="p">
                    Workflow not found.
                </Alert>
            ) : (
                <EditForm workflow={workflow} telemetryRecorder={telemetryRecorder} />
            )}
        </WorkflowPage>
    )
}

/**
 * Form to edit a workflow.
 */
const EditForm: FunctionComponent<TelemetryV2Props & { workflow: WorkflowFields }> = ({
    workflow,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('workflows.update', 'view', {
            metadata: namespaceTelemetryMetadata(workflow.owner),
        })
    }, [telemetryRecorder, workflow.owner])

    const [updateWorkflow, { loading: updateLoading, error: updateError }] = useMutation<
        UpdateWorkflowResult,
        UpdateWorkflowVariables
    >(updateWorkflowMutation, {})

    const navigate = useNavigate()
    const onSubmit = useCallback(
        async (fields: WorkflowFormValue): Promise<void> => {
            try {
                const { data } = await updateWorkflow({
                    variables: {
                        id: workflow.id,
                        input: {
                            name: fields.name,
                            description: fields.description,
                            templateText: fields.templateText,
                            draft: fields.draft,
                        },
                    },
                })
                const updated = data?.updateWorkflow
                if (!updated) {
                    return
                }
                telemetryRecorder.recordEvent('workflows', 'update', {
                    metadata: namespaceTelemetryMetadata(workflow.owner),
                })
                screenReaderAnnounce(`Updated workflow: ${updated.description}`)
                navigate(updated.url, { state: { [WORKFLOW_UPDATED_LOCATION_STATE_KEY]: true } })
            } catch {
                // Mutation error is read in useMutation call.
            }
        },
        [workflow.id, workflow.owner, telemetryRecorder, updateWorkflow, navigate]
    )

    const [showTransferOwnershipModal, setShowTransferOwnershipModal] = useState(false)

    const [deleteWorkflow, { loading: deleteLoading, error: deleteError }] = useMutation(deleteWorkflowMutation)
    const onDeleteClick = useCallback(async (): Promise<void> => {
        if (!workflow) {
            return
        }
        if (!window.confirm(`Delete the workflow ${JSON.stringify(workflow.description)}?`)) {
            return
        }
        try {
            await deleteWorkflow({ variables: { id: workflow.id } })
            telemetryRecorder.recordEvent('workflows', 'delete', {
                metadata: namespaceTelemetryMetadata(workflow.owner),
            })
            navigate(PageRoutes.Workflows)
        } catch (error) {
            logger.error(error)
        }
    }, [workflow, deleteWorkflow, telemetryRecorder, navigate])

    const location = useLocation()

    // Flash after transferring ownership.
    const justTransferredOwnership = !!location.state?.[WORKFLOW_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY]
    useEffect(() => {
        if (justTransferredOwnership) {
            setTimeout(() => navigate({}, { state: {} }), 1000)
        }
    }, [justTransferredOwnership, navigate])

    return (
        <>
            <WorkflowForm
                submitLabel="Save"
                onSubmit={onSubmit}
                otherButtons={
                    <>
                        <div className="flex-grow-1" />
                        {workflow.viewerCanAdminister && (
                            <Button
                                onClick={() => {
                                    telemetryRecorder.recordEvent('workflows.transferOwnership', 'openModal', {
                                        metadata: namespaceTelemetryMetadata(workflow.owner),
                                    })
                                    setShowTransferOwnershipModal(true)
                                }}
                                disabled={updateLoading || deleteLoading}
                                variant="secondary"
                            >
                                Transfer ownership
                            </Button>
                        )}
                        {workflow.viewerCanAdminister && (
                            <Button
                                onClick={onDeleteClick}
                                disabled={updateLoading || deleteLoading}
                                variant="danger"
                                outline={true}
                            >
                                Delete
                            </Button>
                        )}
                    </>
                }
                initialValue={workflow}
                namespaceField={
                    <NamespaceSelector
                        namespaces={[workflow.owner]}
                        disabled={true}
                        label="Owner"
                        className="w-fit-content"
                    />
                }
                loading={updateLoading || deleteLoading}
                error={updateError ?? deleteError}
                flash={justTransferredOwnership ? 'Transferred ownership.' : undefined}
                telemetryRecorder={telemetryRecorder}
            />
            {showTransferOwnershipModal && (
                <TransferOwnershipModal
                    workflow={workflow}
                    onClose={() => {
                        setShowTransferOwnershipModal(false)
                        telemetryRecorder.recordEvent('workflows.transferOwnership', 'closeModal', {
                            metadata: namespaceTelemetryMetadata(workflow.owner),
                        })
                    }}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
        </>
    )
}

const TransferOwnershipModal: FunctionComponent<{
    workflow: Pick<WorkflowFields, 'id' | 'owner'>
    onClose: () => void
    telemetryRecorder: TelemetryRecorder
}> = ({ workflow, onClose, telemetryRecorder }) => {
    const navigate = useNavigate()

    const { namespaces, loading: namespacesLoading, error: namespacesError } = useAffiliatedNamespaces()
    const validNamespaces = namespaces?.filter(ns => ns.id !== workflow.owner.id)
    const [selectedNamespace, setSelectedNamespace] = useState<string | undefined>()
    const selectedNamespaceOrInitial = selectedNamespace ?? validNamespaces?.at(0)?.id

    const [transferWorkflowOwnership, { loading: transferLoading, error: transferError }] = useMutation<
        TransferWorkflowOwnershipResult,
        TransferWorkflowOwnershipVariables
    >(transferWorkflowOwnershipMutation, {})
    const onSubmit = useCallback<FormEventHandler>(
        async (event): Promise<void> => {
            event.preventDefault()
            try {
                const { data } = await transferWorkflowOwnership({
                    variables: { id: workflow.id, newOwner: selectedNamespaceOrInitial! },
                })
                const updated = data?.transferWorkflowOwnership
                if (!updated) {
                    return
                }
                telemetryRecorder.recordEvent('workflows.transferOwnership', 'submit', {
                    metadata: {
                        [`fromNamespaceType${workflow.owner.__typename}`]: 1,
                        [`toNamespaceType${updated.owner.__typename}`]: 1,
                    },
                })
                navigate(`${updated.url}/edit`, {
                    state: { [WORKFLOW_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY]: true },
                })
                onClose()
            } catch (error) {
                logger.error(error)
            }
        },
        [
            transferWorkflowOwnership,
            workflow.id,
            workflow.owner.__typename,
            selectedNamespaceOrInitial,
            telemetryRecorder,
            navigate,
            onClose,
        ]
    )

    const MODAL_LABEL_ID = 'transfer-ownership-modal-label'

    const loading = namespacesLoading || transferLoading

    return (
        <Modal aria-labelledby={MODAL_LABEL_ID} onDismiss={onClose}>
            <Form onSubmit={onSubmit} className="d-flex flex-column flex-gap-4">
                <H3 id={MODAL_LABEL_ID}>Transfer ownership of workflow</H3>
                {namespacesError ? (
                    <ErrorAlert error={namespacesError} />
                ) : loading ? (
                    <LoadingSpinner />
                ) : validNamespaces && validNamespaces.length > 0 && selectedNamespaceOrInitial ? (
                    <>
                        <NamespaceSelector
                            namespaces={validNamespaces}
                            value={selectedNamespaceOrInitial}
                            onSelect={namespace => setSelectedNamespace(namespace)}
                            disabled={transferLoading}
                            loading={namespacesLoading}
                            label="New owner"
                        />
                        <div className="d-flex flex-gap-2">
                            <Button type="submit" disabled={loading} variant="primary">
                                Transfer ownership
                            </Button>
                            <Button onClick={onClose} disabled={loading} variant="secondary" outline={true}>
                                Cancel
                            </Button>
                        </div>
                        {transferError && !loading && <ErrorAlert error={transferError} />}
                    </>
                ) : (
                    <Alert variant="warning">
                        You aren't in any organizations to which you can transfer this workflow.
                    </Alert>
                )}
            </Form>
        </Modal>
    )
}

export const WORKFLOW_UPDATED_LOCATION_STATE_KEY = 'workflow.updated'
const WORKFLOW_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY = 'workflow.transferredOwnership'
