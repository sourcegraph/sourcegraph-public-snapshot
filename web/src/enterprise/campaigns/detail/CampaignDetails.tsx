import slugify from 'slugify'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { UserAvatar } from '../../../user/UserAvatar'
import { Timestamp } from '../../../components/time/Timestamp'
import { CampaignsIcon } from '../icons'
import { noop } from 'lodash'
import { Form } from '../../../components/Form'
import {
    fetchCampaignById,
    updateCampaign,
    deleteCampaign,
    createCampaign,
    fetchCampaignPlanById,
    retryCampaign,
    closeCampaign,
    publishCampaign,
} from './backend'
import { useError, useObservable } from '../../../util/useObservable'
import { asError } from '../../../../../shared/src/util/errors'
import * as H from 'history'
import { CampaignBurndownChart } from './BurndownChart'
import { AddChangesetForm } from './AddChangesetForm'
import { Subject, of, merge, Observable } from 'rxjs'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { ErrorAlert } from '../../../components/alerts'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { Link } from '../../../../../shared/src/components/Link'
import { switchMap, tap, takeWhile, repeatWhen, delay } from 'rxjs/operators'
import { ThemeProps } from '../../../../../shared/src/theme'
import classNames from 'classnames'
import { CampaignTitleField } from './form/CampaignTitleField'
import { CampaignDescriptionField } from './form/CampaignDescriptionField'
import { CloseDeleteCampaignPrompt } from './form/CloseDeleteCampaignPrompt'
import { CampaignStatus } from './CampaignStatus'
import { CampaignTabs } from './CampaignTabs'
import { DEFAULT_CHANGESET_LIST_COUNT } from './presentation'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'

interface Campaign
    extends Pick<
        GQL.ICampaign,
        | '__typename'
        | 'id'
        | 'name'
        | 'description'
        | 'author'
        | 'changesetCountsOverTime'
        | 'branch'
        | 'createdAt'
        | 'updatedAt'
        | 'publishedAt'
        | 'closedAt'
        | 'viewerCanAdminister'
    > {
    plan: Pick<GQL.ICampaignPlan, 'id'> | null
    changesets: Pick<GQL.ICampaign['changesets'], 'nodes' | 'totalCount'>
    changesetPlans: Pick<GQL.ICampaign['changesetPlans'], 'nodes' | 'totalCount'>
    status: Pick<GQL.ICampaign['status'], 'completedCount' | 'pendingCount' | 'errors' | 'state'>
}

interface CampaignPlan extends Pick<GQL.ICampaignPlan, '__typename' | 'id'> {
    changesetPlans: Pick<GQL.ICampaignPlan['changesetPlans'], 'nodes' | 'totalCount'>
    status: Pick<GQL.ICampaignPlan['status'], 'completedCount' | 'pendingCount' | 'errors' | 'state'>
}

interface Props extends ThemeProps {
    /**
     * The campaign ID.
     * If not given, will display a creation form.
     */
    campaignID?: GQL.ID
    authenticatedUser: Pick<GQL.IUser, 'id' | 'username' | 'avatarURL'>
    history: H.History
    location: H.Location

    /** For testing only. */
    _fetchCampaignById?: typeof fetchCampaignById | ((campaign: GQL.ID) => Observable<Campaign | null>)
    _fetchCampaignPlanById?: typeof fetchCampaignPlanById | ((campaignPlan: GQL.ID) => Observable<CampaignPlan | null>)
}

/**
 * The area for a single campaign.
 */
export const CampaignDetails: React.FunctionComponent<Props> = ({
    campaignID,
    history,
    location,
    authenticatedUser,
    isLightTheme,
    _fetchCampaignById = fetchCampaignById,
    _fetchCampaignPlanById = fetchCampaignPlanById,
}) => {
    // State for the form in editing mode
    const [name, setName] = useState<string>('')
    const [description, setDescription] = useState<string>('')
    const [branch, setBranch] = useState<string | null>(null)

    // For errors during fetching
    const triggerError = useError()

    const campaignUpdates = useMemo(() => new Subject<void>(), [])
    const changesetUpdates = useMemo(() => new Subject<void>(), [])

    // Fetch campaign if ID was given
    const [campaign, setCampaign] = useState<Campaign | CampaignPlan | null>()
    useEffect(() => {
        if (!campaignID) {
            return
        }
        const subscription = merge(of(undefined), campaignUpdates)
            .pipe(
                switchMap(
                    () =>
                        new Observable<Campaign | null>(observer => {
                            let currentCampaign: Campaign | null
                            const subscription = _fetchCampaignById(campaignID)
                                .pipe(
                                    tap(campaign => {
                                        currentCampaign = campaign
                                    }),
                                    repeatWhen(obs =>
                                        obs.pipe(
                                            // todo(a8n): why does this not unsubscribe when takeWhile is in outer pipe
                                            takeWhile(
                                                () =>
                                                    currentCampaign?.status?.state ===
                                                    GQL.BackgroundProcessState.PROCESSING
                                            ),
                                            delay(2000)
                                        )
                                    )
                                )
                                .subscribe(observer)
                            return subscription
                        })
                )
            )
            .subscribe({
                next: fetchedCampaign => {
                    setCampaign(fetchedCampaign)
                    changesetUpdates.next()
                },
                error: triggerError,
            })
        return () => subscription.unsubscribe()
    }, [campaignID, triggerError, changesetUpdates, campaignUpdates, _fetchCampaignById])

    const [mode, setMode] = useState<'viewing' | 'editing' | 'saving' | 'deleting' | 'closing'>(
        campaignID ? 'viewing' : 'editing'
    )

    // To report errors from saving or deleting
    const [alertError, setAlertError] = useState<Error>()

    // To unblock the history after leaving edit mode
    const unblockHistoryRef = useRef<H.UnregisterCallback>(noop)
    useEffect(() => {
        if (!campaignID) {
            unblockHistoryRef.current()
            unblockHistoryRef.current = history.block('Do you want to discard this campaign?')
        }
        // Note: the current() method gets dynamically reassigned,
        // therefor we can't return it directly.
        return () => unblockHistoryRef.current()
    }, [campaignID, history])

    const previewCampaignPlans = useMemo(() => new Subject<GQL.ID>(), [])
    const nextPreviewCampaignPlan = useCallback(previewCampaignPlans.next.bind(previewCampaignPlans), [
        previewCampaignPlans,
    ])
    useObservable(
        useMemo(
            () =>
                previewCampaignPlans.pipe(
                    tap(() => {
                        setAlertError(undefined)
                        setCampaign(undefined)
                    }),
                    switchMap(plan => _fetchCampaignPlanById(plan)),
                    tap(campaign => {
                        setCampaign(campaign)
                        if (campaign && campaign.changesetPlans.totalCount <= DEFAULT_CHANGESET_LIST_COUNT) {
                            changesetUpdates.next()
                        }
                    })
                ),
            [previewCampaignPlans, changesetUpdates, _fetchCampaignPlanById]
        )
    )

    const planID: GQL.ID | null = new URLSearchParams(location.search).get('plan')
    useEffect(() => {
        if (planID) {
            nextPreviewCampaignPlan(planID)
        }
    }, [nextPreviewCampaignPlan, planID])

    if (campaign === undefined && campaignID) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    if (campaign === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }

    const specifyingBranchAllowed =
        campaign &&
        (campaign.__typename === 'CampaignPlan' ||
            (!campaign.publishedAt &&
                campaign.changesets.totalCount === 0 &&
                campaign.status.state !== GQL.BackgroundProcessState.PROCESSING))

    const onDraft: React.FormEventHandler = async event => {
        event.preventDefault()
        setMode('saving')
        try {
            const createdCampaign = await createCampaign({
                name,
                description,
                namespace: authenticatedUser.id,
                plan: campaign && campaign.__typename === 'CampaignPlan' ? campaign.id : undefined,
                branch: specifyingBranchAllowed ? branch ?? slugify(name) : undefined,
                draft: true,
            })
            unblockHistoryRef.current()
            history.push(`/campaigns/${createdCampaign.id}`)
            setMode('viewing')
            setAlertError(undefined)
            campaignUpdates.next()
        } catch (err) {
            setMode('editing')
            setAlertError(asError(err))
        }
    }

    const onPublish = async (): Promise<void> => {
        setMode('saving')
        try {
            await publishCampaign(campaign!.id)
            setMode('viewing')
            setAlertError(undefined)
            campaignUpdates.next()
        } catch (err) {
            setMode('editing')
            setAlertError(asError(err))
        }
    }

    const onSubmit: React.FormEventHandler = async event => {
        event.preventDefault()
        setMode('saving')
        try {
            if (campaignID) {
                setCampaign(
                    await updateCampaign({
                        id: campaignID,
                        name,
                        description,
                        branch: specifyingBranchAllowed ? branch ?? slugify(name) : undefined,
                    })
                )
                unblockHistoryRef.current()
            } else {
                const createdCampaign = await createCampaign({
                    name,
                    description,
                    namespace: authenticatedUser.id,
                    plan: campaign && campaign.__typename === 'CampaignPlan' ? campaign.id : undefined,
                    branch: specifyingBranchAllowed ? branch ?? slugify(name) : undefined,
                })
                unblockHistoryRef.current()
                history.push(`/campaigns/${createdCampaign.id}`)
            }
            setMode('viewing')
            setAlertError(undefined)
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
            const { name, description, branch } = campaign as Campaign
            setName(name)
            setDescription(description)
            setBranch(branch)
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
        setAlertError(undefined)
    }

    const onClose = async (closeChangesets: boolean): Promise<void> => {
        if (!confirm('Are you sure you want to close the campaign?')) {
            return
        }
        setMode('closing')
        try {
            await closeCampaign(campaign!.id, closeChangesets)
            campaignUpdates.next()
        } catch (err) {
            setAlertError(asError(err))
        } finally {
            setMode('viewing')
        }
    }

    const onDelete = async (closeChangesets: boolean): Promise<void> => {
        if (!confirm('Are you sure you want to delete the campaign?')) {
            return
        }
        setMode('deleting')
        try {
            await deleteCampaign(campaign!.id, closeChangesets)
            history.push('/campaigns')
        } catch (err) {
            setAlertError(asError(err))
        } finally {
            setMode('viewing')
        }
    }

    const onRetry = async (): Promise<void> => {
        try {
            await retryCampaign(campaign!.id)
            campaignUpdates.next()
        } catch (err) {
            setAlertError(asError(err))
        }
    }

    const onAddChangeset = (): void => {
        // we also check the campaign.changesets.totalCount, so an update to the campaign is required as well
        campaignUpdates.next()
        changesetUpdates.next()
    }

    const author = campaign && campaign.__typename === 'Campaign' ? campaign.author : authenticatedUser

    return (
        <>
            <PageTitle title={campaign && campaign.__typename === 'Campaign' ? campaign.name : 'New campaign'} />
            <Form onSubmit={onSubmit} onReset={onCancel} className="e2e-campaign-form position-relative">
                <div className="d-flex mb-2">
                    <h2 className="m-0">
                        <CampaignsIcon
                            className={classNames(
                                'icon-inline mr-2',
                                campaign && campaign.__typename === 'Campaign' && !campaign.closedAt
                                    ? 'text-success'
                                    : campaignID
                                    ? 'text-danger'
                                    : 'text-muted'
                            )}
                        />
                        <span>
                            <Link to="/campaigns">Campaigns</Link>
                        </span>
                        <span className="text-muted d-inline-block mx-2">/</span>
                        {mode === 'editing' || mode === 'saving' ? (
                            <CampaignTitleField
                                className="w-auto d-inline-block e2e-campaign-title"
                                value={name}
                                onChange={setName}
                                disabled={mode === 'saving'}
                            />
                        ) : (
                            <span>{campaign && campaign.__typename === 'Campaign' && campaign.name}</span>
                        )}
                    </h2>
                    <span className="flex-grow-1 d-flex justify-content-end align-items-center">
                        {(mode === 'saving' || mode === 'deleting' || mode === 'closing') && (
                            <LoadingSpinner className="mr-2" />
                        )}
                        {campaign &&
                            campaign.__typename === 'Campaign' &&
                            (mode === 'editing' || mode === 'saving' ? (
                                <>
                                    <button type="submit" className="btn btn-primary mr-1" disabled={mode === 'saving'}>
                                        Save
                                    </button>
                                    <button type="reset" className="btn btn-secondary" disabled={mode === 'saving'}>
                                        Cancel
                                    </button>
                                </>
                            ) : (
                                campaign.viewerCanAdminister && (
                                    <>
                                        <button
                                            type="button"
                                            id="e2e-campaign-edit"
                                            className="btn btn-secondary mr-1"
                                            onClick={onEdit}
                                            disabled={mode === 'deleting' || mode === 'closing'}
                                        >
                                            Edit
                                        </button>
                                        {!campaign.closedAt && (
                                            <CloseDeleteCampaignPrompt
                                                summary={
                                                    <span
                                                        className={classNames(
                                                            'btn btn-secondary mr-1 dropdown-toggle',
                                                            campaign.status.state ===
                                                                GQL.BackgroundProcessState.PROCESSING && 'disabled'
                                                        )}
                                                        onClick={event =>
                                                            campaign.status.state ===
                                                                GQL.BackgroundProcessState.PROCESSING &&
                                                            event.preventDefault()
                                                        }
                                                        data-tooltip={
                                                            campaign.status.state ===
                                                            GQL.BackgroundProcessState.PROCESSING
                                                                ? 'Cannot close while campaign is being created'
                                                                : undefined
                                                        }
                                                    >
                                                        Close
                                                    </span>
                                                }
                                                message={
                                                    <p>
                                                        Close campaign <strong>{campaign.name}</strong>?
                                                    </p>
                                                }
                                                changesetsCount={campaign.changesets.totalCount}
                                                buttonText="Close"
                                                onButtonClick={onClose}
                                                buttonClassName="btn-secondary"
                                                buttonDisabled={
                                                    mode === 'deleting' ||
                                                    mode === 'closing' ||
                                                    campaign.status.state === GQL.BackgroundProcessState.PROCESSING
                                                }
                                            />
                                        )}
                                        <CloseDeleteCampaignPrompt
                                            summary={
                                                <span
                                                    className={classNames(
                                                        'btn btn-danger dropdown-toggle',
                                                        campaign.status.state ===
                                                            GQL.BackgroundProcessState.PROCESSING && 'disabled'
                                                    )}
                                                    onClick={event =>
                                                        campaign.status.state ===
                                                            GQL.BackgroundProcessState.PROCESSING &&
                                                        event.preventDefault()
                                                    }
                                                    data-tooltip={
                                                        campaign.status.state === GQL.BackgroundProcessState.PROCESSING
                                                            ? 'Cannot delete while campaign is being created'
                                                            : undefined
                                                    }
                                                >
                                                    Delete
                                                </span>
                                            }
                                            message={
                                                <p>
                                                    Delete campaign <strong>{campaign.name}</strong>?
                                                </p>
                                            }
                                            changesetsCount={campaign.changesets.totalCount}
                                            buttonText="Delete"
                                            onButtonClick={onDelete}
                                            buttonClassName="btn-danger"
                                            buttonDisabled={
                                                mode === 'deleting' ||
                                                mode === 'closing' ||
                                                campaign.status.state === GQL.BackgroundProcessState.PROCESSING
                                            }
                                        />
                                    </>
                                )
                            ))}
                    </span>
                </div>
                {alertError && <ErrorAlert error={alertError} />}
                <div className="card">
                    {campaign && campaign.__typename === 'Campaign' && (
                        <div className="card-header">
                            <strong>
                                <UserAvatar user={author} className="icon-inline" /> {author.username}
                            </strong>{' '}
                            started <Timestamp date={campaign.createdAt} />
                        </div>
                    )}
                    {mode === 'editing' || mode === 'saving' ? (
                        <CampaignDescriptionField
                            value={description}
                            onChange={setDescription}
                            disabled={mode === 'saving'}
                        />
                    ) : (
                        campaign &&
                        campaign.__typename === 'Campaign' && (
                            <div className="card-body">
                                <Markdown dangerousInnerHTML={renderMarkdown(campaign.description)}></Markdown>
                            </div>
                        )
                    )}
                </div>
                {mode === 'editing' && (
                    <p className="ml-1 mb-0">
                        <small>
                            <a rel="noopener noreferrer" target="_blank" href="/help/user/markdown">
                                Markdown supported
                            </a>
                        </small>
                    </p>
                )}
                {(!campaign || (campaign && campaign.__typename === 'CampaignPlan')) && (
                    <>
                        {specifyingBranchAllowed && (
                            <div className="form-group mt-3">
                                <label>
                                    Branch name{' '}
                                    <small>
                                        <InformationOutlineIcon
                                            className="icon-inline"
                                            data-tooltip={
                                                'If a branch with the given name already exists, a fallback name will be created by appending a count. Example: "my-branch-name" becomes "my-branch-name-1".'
                                            }
                                        />
                                    </small>
                                </label>
                                <input
                                    type="text"
                                    className="form-control"
                                    onChange={event => setBranch(event.target.value)}
                                    placeholder="my-awesome-campaign"
                                    value={branch !== null ? branch : slugify(name)}
                                    required={true}
                                    disabled={mode === 'saving'}
                                />
                            </div>
                        )}
                        <div className="mt-3">
                            {campaign && (
                                <button
                                    type="submit"
                                    className="btn btn-secondary mr-1"
                                    onClick={onDraft}
                                    disabled={mode !== 'editing'}
                                >
                                    Create draft
                                </button>
                            )}
                            <button
                                type="submit"
                                className="btn btn-primary"
                                disabled={mode !== 'editing' || campaign?.changesetPlans.totalCount === 0}
                            >
                                Create
                            </button>
                        </div>
                    </>
                )}
            </Form>

            {/* is already created or a plan is available */}
            {campaign && (
                <>
                    <CampaignStatus
                        campaign={campaign}
                        status={campaign.status}
                        onPublish={onPublish}
                        onRetry={onRetry}
                    />

                    {campaign.__typename === 'Campaign' && (
                        <>
                            <h3>Progress</h3>
                            <CampaignBurndownChart
                                changesetCountsOverTime={campaign.changesetCountsOverTime}
                                history={history}
                            />
                            {/* only campaigns that have no plan can add changesets manually */}
                            {!campaign.plan && campaign.viewerCanAdminister && (
                                <AddChangesetForm campaignID={campaign.id} onAdd={onAddChangeset} />
                            )}
                        </>
                    )}

                    {campaign.changesetPlans.totalCount +
                        (campaign.__typename === 'Campaign' ? campaign.changesets.totalCount : 0) >
                    0 ? (
                        <CampaignTabs
                            campaign={campaign}
                            changesetUpdates={changesetUpdates}
                            campaignUpdates={campaignUpdates}
                            persistLines={campaign.__typename === 'Campaign'}
                            history={history}
                            location={location}
                            className="mt-3"
                            isLightTheme={isLightTheme}
                        />
                    ) : (
                        campaign.status.state !== GQL.BackgroundProcessState.PROCESSING && (
                            <p className="mt-3 text-muted">No changesets</p>
                        )
                    )}
                </>
            )}
        </>
    )
}
