import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState, useEffect, useRef, useCallback } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { UserAvatar } from '../../../user/UserAvatar'
import { Timestamp } from '../../../components/time/Timestamp'
import { CampaignsIcon } from '../icons'
import { ChangesetList } from './changesets/ChangesetList'
import {
    changesetStatusColorClasses,
    changesetReviewStateColors,
    changesetStageLabels,
} from './changesets/presentation'
import { Link } from '../../../../../shared/src/components/Link'
import { groupBy, noop, upperFirst } from 'lodash'
import { Form } from '../../../components/Form'
import { fetchCampaignById, updateCampaign, deleteCampaign, createCampaign } from './backend'
import { useError } from '../../../util/useObservable'
import { asError } from '../../../../../shared/src/util/errors'
import * as H from 'history'
import { queryNamespaces } from '../../namespaces/backend'

interface Props {
    /**
     * The campaign ID.
     * If not given, will display a creation form.
     */
    campaignID?: GQL.ID

    authenticatedUser: GQL.IUser
    history: H.History
}

const changesetStages: (GQL.ChangesetState | GQL.ChangesetReviewState)[] = [
    GQL.ChangesetState.MERGED,
    GQL.ChangesetState.CLOSED,
    GQL.ChangesetReviewState.APPROVED,
    GQL.ChangesetReviewState.CHANGES_REQUESTED,
    GQL.ChangesetReviewState.PENDING,
]
const changesetStageColors: Record<GQL.ChangesetReviewState | GQL.ChangesetState, string> = {
    ...changesetReviewStateColors,
    ...changesetStatusColorClasses,
}

/**
 * The area for a single campaign.
 */
export const CampaignDetails: React.FunctionComponent<Props> = ({ campaignID, history, authenticatedUser }) => {
    // State for the form in editing mode
    const [name, setName] = useState<string>('')
    const [description, setDescription] = useState<string>('')
    const [namespace, setNamespace] = useState<GQL.ID>()

    const [namespaces, setNamespaces] = useState<GQL.Namespace[]>()
    const getNamespace = useCallback((): GQL.ID | undefined => namespace || (namespaces && namespaces[0].id), [
        namespace,
        namespaces,
    ])

    // For errors during fetching
    const triggerError = useError()

    useEffect(() => {
        if (campaignID) {
            // The namespace of a campaign can only be set on creation.
            return
        }
        const subscription = queryNamespaces().subscribe({ next: setNamespaces, error: triggerError })
        return () => subscription.unsubscribe()
    }, [campaignID, triggerError])

    // Fetch campaign if ID was given
    const [campaign, setCampaign] = useState<GQL.ICampaign | null>()
    useEffect(() => {
        if (!campaignID) {
            return
        }
        const subscription = fetchCampaignById(campaignID).subscribe({ next: setCampaign, error: triggerError })
        return () => subscription.unsubscribe()
    }, [campaignID, triggerError])

    const [mode, setMode] = useState<'viewing' | 'editing' | 'saving' | 'deleting'>(campaignID ? 'viewing' : 'editing')

    // To report errors from saving or deleting
    const [alertError, setAlertError] = useState<Error>()

    // To unblock the history after leaving edit mode
    const unblockHistoryRef = useRef<H.UnregisterCallback>(noop)
    useEffect(() => {
        if (!campaignID) {
            unblockHistoryRef.current()
            unblockHistoryRef.current = history.block('Do you want to discard this campaign?')
        }
        return unblockHistoryRef.current
    }, [campaignID, history])

    if (campaign === undefined && campaignID) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (campaign === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }

    // Calculate campaign progress
    const changeSetCount = campaign ? campaign.changesets.nodes.length : 0
    const changesetsByStage = groupBy(campaign ? campaign.changesets.nodes : [], changeset =>
        // For open changesets, group by review state
        changeset.state !== GQL.ChangesetState.OPEN ? changeset.state : changeset.reviewState
    )

    const onSubmit: React.FormEventHandler = async event => {
        event.preventDefault()
        setMode('saving')
        try {
            if (campaignID) {
                setCampaign(await updateCampaign({ id: campaignID, name, description }))
                unblockHistoryRef.current()
            } else {
                const createdCampaign = await createCampaign({ name, description, namespace: getNamespace()! })
                unblockHistoryRef.current()
                history.push(`/campaigns/${createdCampaign.id}`)
            }
            setMode('viewing')
        } catch (err) {
            setMode('editing')
            setAlertError(asError(err))
        }
    }

    const discardChangesMessage = 'Do you want to discard your changes?'

    const onEdit: React.MouseEventHandler = event => {
        event.preventDefault()
        unblockHistoryRef.current = history.block(discardChangesMessage)
        {
            const { name, description } = campaign!
            setName(name)
            setDescription(description)
            setMode('editing')
        }
    }

    const onCancel: React.FormEventHandler = event => {
        event.preventDefault()
        if (!confirm(discardChangesMessage)) {
            return
        }
        unblockHistoryRef.current()
        setMode('viewing')
    }

    const onDelete: React.MouseEventHandler = async event => {
        event.preventDefault()
        if (!confirm('Are you sure you want to delete the campaign?')) {
            return
        }
        setMode('deleting')
        try {
            await deleteCampaign(campaign!.id)
            history.push('/campaigns')
        } catch (err) {
            setMode('viewing')
            setAlertError(asError(err))
        }
    }

    const author = campaign ? campaign.author : authenticatedUser

    return (
        <>
            <PageTitle title={campaign ? campaign.name : 'New Campaign'} />
            <Form onSubmit={onSubmit} onReset={onCancel}>
                <h2 className="d-flex">
                    <CampaignsIcon className="icon-inline" />
                    {/* The namespace of a campaign can only be set on creation */}
                    {campaign ? (
                        <span>{campaign.namespace.namespaceName}</span>
                    ) : (
                        <select
                            disabled={!namespaces}
                            id="new-campaign-page__namespace"
                            className="form-control w-auto"
                            required={true}
                            value={getNamespace()}
                            onChange={event => setNamespace(event.target.value)}
                        >
                            {namespaces &&
                                namespaces.map(namespace => (
                                    <option value={namespace.id} key={namespace.id}>
                                        {namespace.namespaceName}
                                    </option>
                                ))}
                        </select>
                    )}
                    <span className="text-muted d-inline-block mx-2">/</span>
                    {mode === 'editing' || mode === 'saving' ? (
                        <input
                            className="form-control w-auto d-inline-block"
                            value={name}
                            onChange={event => setName(event.target.value)}
                            placeholder="Campaign title"
                            disabled={mode === 'saving'}
                            autoFocus={true}
                            required={true}
                        />
                    ) : (
                        <span>{campaign!.name}</span>
                    )}
                    <span className="flex-grow-1 d-flex justify-content-end align-items-center">
                        {(mode === 'saving' || mode === 'deleting') && <LoadingSpinner className="mr-2" />}
                        {mode === 'editing' || mode === 'saving' ? (
                            <>
                                <button type="submit" className="btn btn-primary mr-1" disabled={mode === 'saving'}>
                                    Save
                                </button>
                                {campaignID && (
                                    <button type="reset" className="btn btn-secondary" disabled={mode === 'saving'}>
                                        Cancel
                                    </button>
                                )}
                            </>
                        ) : (
                            <>
                                <button
                                    type="button"
                                    className="btn btn-secondary mr-1"
                                    onClick={onEdit}
                                    disabled={mode === 'deleting'}
                                >
                                    Edit
                                </button>
                                <button
                                    type="button"
                                    className="btn btn-danger"
                                    onClick={onDelete}
                                    disabled={mode === 'deleting'}
                                >
                                    Delete
                                </button>
                            </>
                        )}
                    </span>
                </h2>
                {alertError && <div className="alert alert-danger">{upperFirst(alertError.message)}</div>}
                <div className="card mb-3">
                    <div className="card-header">
                        <strong>
                            <UserAvatar user={author} className="icon-inline" /> {author.username}
                        </strong>
                        {campaign && (
                            <>
                                {' '}
                                started <Timestamp date={campaign.createdAt} />
                            </>
                        )}
                    </div>
                    {mode === 'editing' || mode === 'saving' ? (
                        <textarea
                            className="form-control"
                            value={description}
                            onChange={event => setDescription(event.target.value)}
                            placeholder="Describe the purpose of this campaign, link to relevant internal documentation, etc."
                            disabled={mode === 'saving'}
                        />
                    ) : (
                        <div className="card-body">{campaign!.description}</div>
                    )}
                </div>
            </Form>
            {campaign && (
                <>
                    <h3>
                        Changesets{' '}
                        <span className="badge badge-secondary badge-pill">{campaign.changesets.nodes.length}</span>
                    </h3>
                    {changeSetCount > 0 && (
                        <div>
                            <div className="progress rounded mb-2">
                                {changesetStages.map(stage => {
                                    const changesetsInStage = changesetsByStage[stage] || []
                                    const count = changesetsInStage.length
                                    return (
                                        count > 0 && (
                                            <div
                                                // Needed for dynamic width
                                                // eslint-disable-next-line react/forbid-dom-props
                                                style={{ width: (count / changeSetCount) * 100 + '%' }}
                                                className={`progress-bar bg-${changesetStageColors[stage]}`}
                                                role="progressbar"
                                                aria-valuemin={0}
                                                aria-valuenow={count}
                                                aria-valuemax={changeSetCount}
                                                key={stage}
                                            >
                                                {count} {changesetStageLabels[stage]}
                                            </div>
                                        )
                                    )
                                })}
                            </div>
                        </div>
                    )}
                    <ChangesetList changesets={campaign.changesets.nodes} />
                    <p className="mt-2">
                        Use the <Link to="/api/console">GraphQL API</Link> to add changesets to this campaign (
                        <code>createChangesets</code> and <code>addChangesetsToCampaign</code>)
                    </p>
                </>
            )}
        </>
    )
}
