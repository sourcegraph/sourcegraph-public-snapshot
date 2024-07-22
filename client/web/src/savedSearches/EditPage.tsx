import { useCallback, useEffect, useState, type FormEventHandler, type FunctionComponent } from 'react'

import { mdiLink } from '@mdi/js'
import { truncate } from 'lodash'
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
    Icon,
    Label,
    Link,
    LoadingSpinner,
    Modal,
    PageHeader,
    screenReaderAnnounce,
} from '@sourcegraph/wildcard'

import {
    SavedSearchVisibility,
    type ChangeSavedSearchVisibilityResult,
    type ChangeSavedSearchVisibilityVariables,
    type SavedSearchFields,
    type SavedSearchResult,
    type SavedSearchVariables,
    type TransferSavedSearchOwnershipResult,
    type TransferSavedSearchOwnershipVariables,
    type UpdateSavedSearchResult,
    type UpdateSavedSearchVariables,
} from '../graphql-operations'
import { LibraryItemVisibilityBadge } from '../library/itemBadges'
import { useLibraryConfiguration } from '../library/useLibraryConfiguration'
import { NamespaceSelector } from '../namespaces/NamespaceSelector'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'

import { SavedSearchForm, type SavedSearchFormValue } from './Form'
import {
    changeSavedSearchVisibilityMutation,
    deleteSavedSearchMutation,
    savedSearchQuery,
    transferSavedSearchOwnershipMutation,
    updateSavedSearchMutation,
} from './graphql'
import { SavedSearchPage } from './Page'
import { urlToEditSavedSearch } from './util'

/**
 * Page to edit a saved search.
 */
export const EditPage: FunctionComponent<{ isSourcegraphDotCom: boolean } & TelemetryV2Props> = ({
    isSourcegraphDotCom,
    telemetryRecorder,
}) => {
    const { id } = useParams<{ id: string }>()
    const { data, loading, error } = useQuery<SavedSearchResult, SavedSearchVariables>(savedSearchQuery, {
        variables: { id: id! },
    })
    const savedSearch = data?.node?.__typename === 'SavedSearch' ? data.node : null

    return (
        <SavedSearchPage
            title={savedSearch ? `Editing: ${savedSearch.description} - saved search` : 'Edit saved search'}
            actions={
                savedSearch && (
                    <Button to={savedSearch.url} variant="secondary" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiLink} /> Permalink
                    </Button>
                )
            }
            breadcrumbsNamespace={savedSearch?.owner}
            breadcrumbs={
                savedSearch ? (
                    <>
                        <PageHeader.Breadcrumb to={savedSearch.url}>
                            {truncate(savedSearch.description, { length: 30 })}
                        </PageHeader.Breadcrumb>
                        <PageHeader.Breadcrumb>Edit</PageHeader.Breadcrumb>
                    </>
                ) : null
            }
        >
            {loading ? (
                <LoadingSpinner />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : !savedSearch ? (
                <Alert variant="danger" as="p">
                    Saved search not found.
                </Alert>
            ) : !savedSearch.viewerCanAdminister ? (
                <Alert variant="danger" as="p">
                    You do not have permission to edit this saved search.
                </Alert>
            ) : (
                <EditForm
                    savedSearch={savedSearch}
                    isSourcegraphDotCom={isSourcegraphDotCom}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
        </SavedSearchPage>
    )
}

/**
 * Form to edit a saved search.
 */
const EditForm: FunctionComponent<
    TelemetryV2Props & { savedSearch: SavedSearchFields; isSourcegraphDotCom: boolean }
> = ({ savedSearch, telemetryRecorder, isSourcegraphDotCom }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('savedSearches.update', 'view', {
            metadata: namespaceTelemetryMetadata(savedSearch.owner),
        })
    }, [telemetryRecorder, savedSearch.owner])

    const location = useLocation()

    const [updateSavedSearch, { loading: updateLoading, error: updateError }] = useMutation<
        UpdateSavedSearchResult,
        UpdateSavedSearchVariables
    >(updateSavedSearchMutation, {})

    const navigate = useNavigate()
    const onSubmit = useCallback(
        async (fields: SavedSearchFormValue): Promise<void> => {
            try {
                const { data } = await updateSavedSearch({
                    variables: {
                        id: savedSearch.id,
                        input: {
                            description: fields.description,
                            query: fields.query,
                            draft: fields.draft,
                        },
                    },
                })
                const updated = data?.updateSavedSearch
                if (!updated) {
                    return
                }
                telemetryRecorder.recordEvent('savedSearches', 'update', {
                    metadata: namespaceTelemetryMetadata(savedSearch.owner),
                })
                screenReaderAnnounce(`Updated saved search: ${updated.description}`)
                navigate(updated.url, { state: { [SAVED_SEARCH_UPDATED_LOCATION_STATE_KEY]: true } })
            } catch {
                // Mutation error is read in useMutation call.
            }
        },
        [savedSearch.id, savedSearch.owner, telemetryRecorder, updateSavedSearch, navigate]
    )

    const [showTransferOwnershipModal, setShowTransferOwnershipModal] = useState(false)

    const [deleteSavedSearch, { loading: deleteLoading, error: deleteError }] = useMutation(deleteSavedSearchMutation)
    const onDeleteClick = useCallback(async (): Promise<void> => {
        if (!savedSearch) {
            return
        }
        if (!window.confirm(`Delete the saved search ${JSON.stringify(savedSearch.description)}?`)) {
            return
        }
        try {
            await deleteSavedSearch({ variables: { id: savedSearch.id } })
            telemetryRecorder.recordEvent('savedSearches', 'delete', {
                metadata: namespaceTelemetryMetadata(savedSearch.owner),
            })
            navigate(PageRoutes.SavedSearches)
        } catch (error) {
            logger.error(error)
        }
    }, [savedSearch, deleteSavedSearch, telemetryRecorder, navigate])

    const [changeSavedSearchVisibility, { loading: changeVisibilityLoading, error: changeVisibilityError }] =
        useMutation<ChangeSavedSearchVisibilityResult, ChangeSavedSearchVisibilityVariables>(
            changeSavedSearchVisibilityMutation
        )
    const onChangeVisibilityClick = useCallback(async (): Promise<void> => {
        if (!savedSearch) {
            return
        }
        try {
            const newVisibility =
                savedSearch.visibility === SavedSearchVisibility.PUBLIC
                    ? SavedSearchVisibility.SECRET
                    : SavedSearchVisibility.PUBLIC
            await changeSavedSearchVisibility({
                variables: {
                    id: savedSearch.id,
                    newVisibility,
                },
            })
            telemetryRecorder.recordEvent('savedSearches', 'changeVisibility', {
                metadata: {
                    ...namespaceTelemetryMetadata(savedSearch.owner),
                    toVisibilityPublic: newVisibility === SavedSearchVisibility.PUBLIC ? 1 : 0,
                    toVisibilitySecret: newVisibility === SavedSearchVisibility.SECRET ? 1 : 0,
                },
            })
            screenReaderAnnounce(`Changed visibility of saved search: ${JSON.stringify(savedSearch.description)}`)
            navigate(location, {
                replace: true,
                state: { [SAVED_SEARCH_CHANGED_VISIBILITY_LOCATION_STATE_KEY]: true },
            })
        } catch (error) {
            logger.error(error)
        }
    }, [savedSearch, changeSavedSearchVisibility, telemetryRecorder, navigate, location])

    // Flash after changing visibility.
    const justChangedVisibility = !!location.state?.[SAVED_SEARCH_CHANGED_VISIBILITY_LOCATION_STATE_KEY]
    useEffect(() => {
        if (justChangedVisibility) {
            setTimeout(() => navigate({}, { state: {} }), 1000)
        }
    }, [justChangedVisibility, navigate])

    // Flash after transferring ownership.
    const justTransferredOwnership = !!location.state?.[SAVED_SEARCH_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY]
    useEffect(() => {
        if (justTransferredOwnership) {
            setTimeout(() => navigate({}, { state: {} }), 1000)
        }
    }, [justTransferredOwnership, navigate])

    const { viewerCanChangeLibraryItemVisibilityToPublic } = useLibraryConfiguration()

    return (
        <>
            <SavedSearchForm
                submitLabel="Save"
                onSubmit={onSubmit}
                otherButtons={
                    <Button to={savedSearch.url} variant="secondary" outline={true} as={Link}>
                        Cancel
                    </Button>
                }
                isSourcegraphDotCom={isSourcegraphDotCom}
                initialValue={savedSearch}
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
                beforeFields={
                    <NamespaceSelector
                        namespaces={[savedSearch.owner]}
                        disabled={true}
                        label="Owner"
                        className="w-fit-content mb-4"
                    />
                }
                afterFields={
                    <div className="form-group">
                        <Label className="mr-2">Visibility:</Label>
                        <LibraryItemVisibilityBadge item={savedSearch} />
                    </div>
                }
            />

            <H4 className="mt-5 mb-1">Other actions</H4>
            <Container className="d-inline-block">
                <div className="d-flex flex-column align-items-start flex-gap-4">
                    {savedSearch.viewerCanAdminister && (
                        <Button
                            onClick={() => {
                                telemetryRecorder.recordEvent('savedSearches.transferOwnership', 'openModal', {
                                    metadata: namespaceTelemetryMetadata(savedSearch.owner),
                                })
                                setShowTransferOwnershipModal(true)
                            }}
                            disabled={updateLoading || deleteLoading || changeVisibilityLoading}
                            variant="secondary"
                        >
                            Transfer ownership
                        </Button>
                    )}
                    {savedSearch.viewerCanAdminister && viewerCanChangeLibraryItemVisibilityToPublic && (
                        <Button
                            onClick={onChangeVisibilityClick}
                            disabled={updateLoading || deleteLoading || changeVisibilityLoading}
                            variant="warning"
                            outline={true}
                        >
                            Change visibility to{' '}
                            {savedSearch.visibility === SavedSearchVisibility.PUBLIC ? 'secret' : 'public'}
                        </Button>
                    )}
                    {savedSearch.viewerCanAdminister && (
                        <Button
                            onClick={onDeleteClick}
                            disabled={updateLoading || deleteLoading || changeVisibilityLoading}
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
                    savedSearch={savedSearch}
                    onClose={() => {
                        setShowTransferOwnershipModal(false)
                        telemetryRecorder.recordEvent('savedSearches.transferOwnership', 'closeModal', {
                            metadata: namespaceTelemetryMetadata(savedSearch.owner),
                        })
                    }}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
        </>
    )
}

const TransferOwnershipModal: FunctionComponent<{
    savedSearch: Pick<SavedSearchFields, 'id' | 'owner'>
    onClose: () => void
    telemetryRecorder: TelemetryRecorder
}> = ({ savedSearch, onClose, telemetryRecorder }) => {
    const navigate = useNavigate()

    const { namespaces, loading: namespacesLoading, error: namespacesError } = useAffiliatedNamespaces()
    const validNamespaces = namespaces?.filter(ns => ns.id !== savedSearch.owner.id)
    const [selectedNamespace, setSelectedNamespace] = useState<string | undefined>()
    const selectedNamespaceOrInitial = selectedNamespace ?? validNamespaces?.at(0)?.id

    const [transferSavedSearchOwnership, { loading: transferLoading, error: transferError }] = useMutation<
        TransferSavedSearchOwnershipResult,
        TransferSavedSearchOwnershipVariables
    >(transferSavedSearchOwnershipMutation, {})
    const onSubmit = useCallback<FormEventHandler>(
        async (event): Promise<void> => {
            event.preventDefault()
            try {
                const { data } = await transferSavedSearchOwnership({
                    variables: { id: savedSearch.id, newOwner: selectedNamespaceOrInitial! },
                })
                const updated = data?.transferSavedSearchOwnership
                if (!updated) {
                    return
                }
                telemetryRecorder.recordEvent('savedSearches.transferOwnership', 'submit', {
                    metadata: {
                        [`fromNamespaceType${savedSearch.owner.__typename}`]: 1,
                        [`toNamespaceType${updated.owner.__typename}`]: 1,
                    },
                })
                navigate(urlToEditSavedSearch(updated), {
                    state: { [SAVED_SEARCH_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY]: true },
                })
                onClose()
            } catch (error) {
                logger.error(error)
            }
        },
        [
            transferSavedSearchOwnership,
            savedSearch.id,
            savedSearch.owner.__typename,
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
                <H3 id={MODAL_LABEL_ID}>Transfer ownership of saved search</H3>
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
                        You aren't in any organizations to which you can transfer this saved search.
                    </Alert>
                )}
            </Form>
        </Modal>
    )
}

export const SAVED_SEARCH_UPDATED_LOCATION_STATE_KEY = 'savedSearch.updated'
const SAVED_SEARCH_TRANSFERRED_OWNERSHIP_LOCATION_STATE_KEY = 'savedSearch.transferredOwnership'
const SAVED_SEARCH_CHANGED_VISIBILITY_LOCATION_STATE_KEY = 'savedSearch.changedVisibility'
