import { useCallback, useEffect, useMemo, useState, type FormEventHandler, type FunctionComponent } from 'react'

import { useLocation, useNavigate, useParams } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import type { TelemetryRecorder, TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Alert,
    Button,
    Container,
    ErrorAlert,
    Form,
    H3,
    H4,
    Label,
    Link,
    LoadingSpinner,
    Modal,
    PageHeader,
    screenReaderAnnounce,
} from '@sourcegraph/wildcard'

import {
    PromptVisibility,
    type ChangePromptVisibilityResult,
    type ChangePromptVisibilityVariables,
    type PromptFields,
    type PromptResult,
    type PromptVariables,
    type TransferPromptOwnershipResult,
    type TransferPromptOwnershipVariables,
    type UpdatePromptResult,
    type UpdatePromptVariables,
} from '../graphql-operations'
import { LibraryItemVisibilityBadge } from '../library/itemBadges'
import { useLibraryConfiguration } from '../library/useLibraryConfiguration'
import { NamespaceSelector } from '../namespaces/NamespaceSelector'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'

import { PromptForm, type PromptFormValue } from './Form'
import {
    changePromptVisibilityMutation,
    deletePromptMutation,
    promptQuery,
    transferPromptOwnershipMutation,
    updatePromptMutation,
} from './graphql'
import { PromptPage } from './Page'
import { urlToEditPrompt } from './util'

/**
 * Page to edit a prompt.
 */
export const EditPage: FunctionComponent<TelemetryV2Props> = ({ telemetryRecorder }) => {
    const { id } = useParams<{ id: string }>()
    const { data, loading, error } = useQuery<PromptResult, PromptVariables>(promptQuery, {
        variables: { id: id! },
    })
    const prompt = data?.node?.__typename === 'Prompt' ? data.node : null

    return (
        <PromptPage
            title={prompt ? `Editing: ${prompt.nameWithOwner} - prompt` : 'Edit prompt'}
            breadcrumbsNamespace={prompt?.owner}
            breadcrumbs={
                prompt ? (
                    <>
                        <PageHeader.Breadcrumb to={prompt.url}>{prompt.name}</PageHeader.Breadcrumb>
                        <PageHeader.Breadcrumb>Edit</PageHeader.Breadcrumb>
                    </>
                ) : null
            }
        >
            {loading ? (
                <LoadingSpinner />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : !prompt ? (
                <Alert variant="danger" as="p">
                    Prompt not found.
                </Alert>
            ) : !prompt.viewerCanAdminister ? (
                <Alert variant="danger" as="p">
                    You do not have permission to edit this prompt.
                </Alert>
            ) : (
                <EditForm prompt={prompt} telemetryRecorder={telemetryRecorder} />
            )}
        </PromptPage>
    )
}

/**
 * Form to edit a prompt.
 */
const EditForm: FunctionComponent<TelemetryV2Props & { prompt: PromptFields }> = ({ prompt, telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('prompts.update', 'view', {
            metadata: namespaceTelemetryMetadata(prompt.owner),
        })
    }, [telemetryRecorder, prompt.owner])

    const location = useLocation()

    const [updatePrompt, { loading: updateLoading, error: updateError }] = useMutation<
        UpdatePromptResult,
        UpdatePromptVariables
    >(updatePromptMutation, {})

    const navigate = useNavigate()
    const onSubmit = useCallback(
        async (fields: PromptFormValue): Promise<void> => {
            try {
                const { data } = await updatePrompt({
                    variables: {
                        id: prompt.id,
                        input: {
                            name: fields.name,
                            description: fields.description,
                            definitionText: fields.definitionText,
                            draft: fields.draft,
                        },
                    },
                })
                const updated = data?.updatePrompt
                if (!updated) {
                    return
                }
                telemetryRecorder.recordEvent('prompts', 'update', {
                    metadata: namespaceTelemetryMetadata(prompt.owner),
                })
                screenReaderAnnounce(`Updated prompt: ${updated.nameWithOwner}`)
                navigate(updated.url, { state: { [PROMPT_UPDATED_LOCATION_STATE_KEY]: true } })
            } catch {
                // Mutation error is read in useMutation call.
            }
        },
        [prompt.id, prompt.owner, telemetryRecorder, updatePrompt, navigate]
    )

    const [showTransferOwnershipModal, setShowTransferOwnershipModal] = useState(false)

    const [deletePrompt, { loading: deleteLoading, error: deleteError }] = useMutation(deletePromptMutation)
    const onDeleteClick = useCallback(async (): Promise<void> => {
        if (!prompt) {
            return
        }
        if (!window.confirm(`Delete the prompt ${JSON.stringify(prompt.nameWithOwner)}?`)) {
            return
        }
        try {
            await deletePrompt({ variables: { id: prompt.id } })
            telemetryRecorder.recordEvent('prompts', 'delete', {
                metadata: namespaceTelemetryMetadata(prompt.owner),
            })
            navigate(PageRoutes.Prompts)
        } catch (error) {
            logger.error(error)
        }
    }, [prompt, deletePrompt, telemetryRecorder, navigate])

    // Flash after transferring ownership.
    const justTransferredOwnership = !!location.state?.[PROMPT_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY]
    useEffect(() => {
        if (justTransferredOwnership) {
            setTimeout(() => navigate({}, { state: {} }), 1000)
        }
    }, [justTransferredOwnership, navigate])

    const [changePromptVisibility, { loading: changeVisibilityLoading, error: changeVisibilityError }] = useMutation<
        ChangePromptVisibilityResult,
        ChangePromptVisibilityVariables
    >(changePromptVisibilityMutation)
    const onChangeVisibilityClick = useCallback(async (): Promise<void> => {
        if (!prompt) {
            return
        }
        try {
            const newVisibility =
                prompt.visibility === PromptVisibility.PUBLIC ? PromptVisibility.SECRET : PromptVisibility.PUBLIC
            await changePromptVisibility({
                variables: {
                    id: prompt.id,
                    newVisibility,
                },
            })
            telemetryRecorder.recordEvent('savedSearches', 'changeVisibility', {
                metadata: {
                    ...namespaceTelemetryMetadata(prompt.owner),
                    toVisibilityPublic: newVisibility === PromptVisibility.PUBLIC ? 1 : 0,
                    toVisibilitySecret: newVisibility === PromptVisibility.SECRET ? 1 : 0,
                },
            })
            screenReaderAnnounce(`Changed visibility of saved search: ${JSON.stringify(prompt.description)}`)
            navigate(location, {
                replace: true,
                state: { [PROMPT_CHANGED_VISIBILITY_LOCATION_STATE_KEY]: true },
            })
        } catch (error) {
            logger.error(error)
        }
    }, [prompt, changePromptVisibility, telemetryRecorder, navigate, location])

    // Flash after changing visibility.
    const justChangedVisibility = !!location.state?.[PROMPT_CHANGED_VISIBILITY_LOCATION_STATE_KEY]
    useEffect(() => {
        if (justChangedVisibility) {
            setTimeout(() => navigate({}, { state: {} }), 1000)
        }
    }, [justChangedVisibility, navigate])

    const { viewerCanChangeLibraryItemVisibilityToPublic } = useLibraryConfiguration()

    const initialValue = useMemo<PromptFormValue>(
        () => ({
            name: prompt.name,
            description: prompt.description,
            definitionText: prompt.definition.text,
            draft: prompt.draft,
        }),
        [prompt]
    )

    return (
        <>
            <PromptForm
                submitLabel="Save"
                onSubmit={onSubmit}
                otherButtons={
                    <Button to={prompt.url} variant="secondary" outline={true} as={Link}>
                        Cancel
                    </Button>
                }
                initialValue={initialValue}
                namespaceField={
                    <NamespaceSelector
                        namespaces={[prompt.owner]}
                        disabled={true}
                        label="Owner"
                        className="w-fit-content"
                    />
                }
                loading={updateLoading || deleteLoading || changeVisibilityLoading}
                error={updateError ?? deleteError ?? changeVisibilityError}
                flash={
                    justTransferredOwnership
                        ? 'Transferred ownership.'
                        : justChangedVisibility
                        ? 'Changed visibility.'
                        : undefined
                }
                telemetryRecorder={telemetryRecorder}
                afterFields={
                    <div className="form-group">
                        <Label className="mr-2">Visibility:</Label>
                        <LibraryItemVisibilityBadge item={prompt} />
                    </div>
                }
            />

            <H4 className="mt-5 mb-1">Other actions</H4>
            <Container className="d-inline-block">
                <div className="d-flex flex-column align-items-start flex-gap-4">
                    {prompt.viewerCanAdminister && (
                        <Button
                            onClick={() => {
                                telemetryRecorder.recordEvent('prompts.transferOwnership', 'openModal', {
                                    metadata: namespaceTelemetryMetadata(prompt.owner),
                                })
                                setShowTransferOwnershipModal(true)
                            }}
                            disabled={updateLoading || deleteLoading}
                            variant="secondary"
                        >
                            Transfer ownership
                        </Button>
                    )}
                    {prompt.viewerCanAdminister && viewerCanChangeLibraryItemVisibilityToPublic && (
                        <Button
                            onClick={onChangeVisibilityClick}
                            disabled={updateLoading || deleteLoading || changeVisibilityLoading}
                            variant="warning"
                            outline={true}
                        >
                            Change visibility to {prompt.visibility === PromptVisibility.PUBLIC ? 'secret' : 'public'}
                        </Button>
                    )}
                    {prompt.viewerCanAdminister && (
                        <Button
                            onClick={onDeleteClick}
                            disabled={updateLoading || deleteLoading}
                            variant="danger"
                            outline={true}
                        >
                            Delete
                        </Button>
                    )}
                </div>
            </Container>

            {showTransferOwnershipModal && (
                <TransferOwnershipModal
                    prompt={prompt}
                    onClose={() => {
                        setShowTransferOwnershipModal(false)
                        telemetryRecorder.recordEvent('prompts.transferOwnership', 'closeModal', {
                            metadata: namespaceTelemetryMetadata(prompt.owner),
                        })
                    }}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
        </>
    )
}

const TransferOwnershipModal: FunctionComponent<{
    prompt: Pick<PromptFields, 'id' | 'owner'>
    onClose: () => void
    telemetryRecorder: TelemetryRecorder
}> = ({ prompt, onClose, telemetryRecorder }) => {
    const navigate = useNavigate()

    const { namespaces, loading: namespacesLoading, error: namespacesError } = useAffiliatedNamespaces()
    const validNamespaces = namespaces?.filter(ns => ns.id !== prompt.owner.id)
    const [selectedNamespace, setSelectedNamespace] = useState<string | undefined>()
    const selectedNamespaceOrInitial = selectedNamespace ?? validNamespaces?.at(0)?.id

    const [transferPromptOwnership, { loading: transferLoading, error: transferError }] = useMutation<
        TransferPromptOwnershipResult,
        TransferPromptOwnershipVariables
    >(transferPromptOwnershipMutation, {})
    const onSubmit = useCallback<FormEventHandler>(
        async (event): Promise<void> => {
            event.preventDefault()
            try {
                const { data } = await transferPromptOwnership({
                    variables: { id: prompt.id, newOwner: selectedNamespaceOrInitial! },
                })
                const updated = data?.transferPromptOwnership
                if (!updated) {
                    return
                }
                telemetryRecorder.recordEvent('prompts.transferOwnership', 'submit', {
                    metadata: {
                        [`fromNamespaceType${prompt.owner.__typename}`]: 1,
                        [`toNamespaceType${updated.owner.__typename}`]: 1,
                    },
                })
                navigate(urlToEditPrompt(updated), {
                    state: { [PROMPT_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY]: true },
                })
                onClose()
            } catch (error) {
                logger.error(error)
            }
        },
        [
            transferPromptOwnership,
            prompt.id,
            prompt.owner.__typename,
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
                <H3 id={MODAL_LABEL_ID}>Transfer ownership of prompt</H3>
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
                        You aren't in any organizations to which you can transfer this prompt.
                    </Alert>
                )}
            </Form>
        </Modal>
    )
}

export const PROMPT_UPDATED_LOCATION_STATE_KEY = 'prompt.updated'
const PROMPT_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY = 'prompt.transferredOwnership'
const PROMPT_CHANGED_VISIBILITY_LOCATION_STATE_KEY = 'prompt.changedVisibility'
