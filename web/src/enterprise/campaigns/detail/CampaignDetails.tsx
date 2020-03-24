import slugify from 'slugify'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState, useEffect, useRef, useMemo, ChangeEvent } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { UserAvatar } from '../../../user/UserAvatar'
import { Timestamp } from '../../../components/time/Timestamp'
import { noop } from 'lodash'
import { Form } from '../../../components/Form'
import {
    fetchCampaignById,
    updateCampaign,
    deleteCampaign,
    createCampaign,
    retryCampaign,
    closeCampaign,
    publishCampaign,
    fetchPatchSetById,
} from './backend'
import { useError, useObservable } from '../../../../../shared/src/util/useObservable'
import { asError } from '../../../../../shared/src/util/errors'
import * as H from 'history'
import { CampaignBurndownChart } from './BurndownChart'
import { AddChangesetForm } from './AddChangesetForm'
import { Subject, of, merge, Observable, NEVER } from 'rxjs'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { ErrorAlert } from '../../../components/alerts'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { switchMap, tap, takeWhile, repeatWhen, delay } from 'rxjs/operators'
import { ThemeProps } from '../../../../../shared/src/theme'
import { CampaignDescriptionField } from './form/CampaignDescriptionField'
import { CampaignStatus } from './CampaignStatus'
import { CampaignUpdateDiff } from './CampaignUpdateDiff'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import { CampaignActionsBar } from './CampaignActionsBar'
import { CampaignTitleField } from './form/CampaignTitleField'
import { CampaignChangesets } from './changesets/CampaignChangesets'
import { CampaignDiffStat } from './CampaignDiffStat'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignPatches } from './patches/CampaignPatches'
import { PatchSetPatches } from './patches/PatchSetPatches'

export type CampaignUIMode = 'viewing' | 'editing' | 'saving' | 'deleting' | 'closing'

interface Campaign
    extends Pick<
        GQL.ICampaign,
        | '__typename'
        | 'id'
        | 'name'
        | 'description'
        | 'author'
        | 'changesetCountsOverTime'
        | 'createdAt'
        | 'updatedAt'
        | 'publishedAt'
        | 'closedAt'
        | 'viewerCanAdminister'
        | 'branch'
    > {
    patchSet: Pick<GQL.IPatchSet, 'id'> | null
    changesets: Pick<GQL.ICampaign['changesets'], 'nodes' | 'totalCount'>
    patches: Pick<GQL.ICampaign['patches'], 'nodes' | 'totalCount'>
    status: Pick<GQL.ICampaign['status'], 'completedCount' | 'pendingCount' | 'errors' | 'state'>
}

interface PatchSet extends Pick<GQL.IPatchSet, '__typename' | 'id'> {
    patches: Pick<GQL.IPatchSet['patches'], 'nodes' | 'totalCount'>
}

interface Props extends ThemeProps, ExtensionsControllerProps, PlatformContextProps, TelemetryProps {
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
    /** For testing only. */
    _fetchPatchSetById?: typeof fetchPatchSetById | ((patchSet: GQL.ID) => Observable<PatchSet | null>)
    /** For testing only. */
    _noSubject?: boolean
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
    extensionsController,
    platformContext,
    telemetryService,
    _fetchCampaignById = fetchCampaignById,
    _fetchPatchSetById = fetchPatchSetById,
    _noSubject = false,
}) => {
    // State for the form in editing mode
    const [name, setName] = useState<string>('')
    const [description, setDescription] = useState<string>('')
    const [branch, setBranch] = useState<string>('')
    const [branchModified, setBranchModified] = useState<boolean>(false)

    // For errors during fetching
    const triggerError = useError()

    /** Retrigger campaign fetching */
    const campaignUpdates = useMemo(() => new Subject<void>(), [])
    /** Retrigger changeset fetching */
    const changesetUpdates = useMemo(() => new Subject<void>(), [])

    const [campaign, setCampaign] = useState<Campaign | null>()
    useEffect(() => {
        if (!campaignID) {
            return
        }
        // on the very first fetch, a reload of the changesets is not required
        let isFirstCampaignFetch = true
        // Fetch campaign if ID was given
        const subscription = merge(of(undefined), _noSubject ? new Observable<void>() : campaignUpdates)
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
                                    // repeat fetching the campaign as long as the state is still processing
                                    repeatWhen(obs =>
                                        obs.pipe(
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
                    if (fetchedCampaign) {
                        setName(fetchedCampaign.name)
                        setDescription(fetchedCampaign.description)
                        setBranch(fetchedCampaign.branch ?? '')
                        setBranchModified(false)
                    }
                    if (!isFirstCampaignFetch) {
                        changesetUpdates.next()
                    }
                    isFirstCampaignFetch = false
                },
                error: triggerError,
            })
        return () => subscription.unsubscribe()
    }, [campaignID, triggerError, changesetUpdates, campaignUpdates, _fetchCampaignById, _noSubject])

    const [mode, setMode] = useState<CampaignUIMode>(campaignID ? 'viewing' : 'editing')

    // To report errors from saving or deleting
    const [alertError, setAlertError] = useState<Error>()

    const patchSetID: GQL.ID | null = new URLSearchParams(location.search).get('patchSet')
    useEffect(() => {
        if (patchSetID) {
            setMode('editing')
        }
    }, [patchSetID])

    // To unblock the history after leaving edit mode
    const unblockHistoryRef = useRef<H.UnregisterCallback>(noop)
    useEffect(() => {
        if (!campaignID && patchSetID === null) {
            unblockHistoryRef.current()
            unblockHistoryRef.current = history.block('Do you want to discard this campaign?')
        }
        // Note: the current() method gets dynamically reassigned,
        // therefor we can't return it directly.
        return () => unblockHistoryRef.current()
    }, [campaignID, history, patchSetID])

    const patchSet = useObservable(
        useMemo(() => (!patchSetID ? NEVER : _fetchPatchSetById(patchSetID)), [patchSetID, _fetchPatchSetById])
    )

    // Is loading
    if ((campaignID && campaign === undefined) || (patchSetID && patchSet === undefined)) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    // Campaign was not found
    if (campaign === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }
    // Patch set was not found
    if (patchSet === null) {
        return <HeroPage icon={AlertCircleIcon} title="Patch set not found" />
    }
    const updateMode = !!campaign && !!patchSet

    if (updateMode && campaign) {
        if (!campaign.patchSet?.id) {
            return <HeroPage icon={AlertCircleIcon} title="Cannot update a manual campaign with a patch set" />
        }
        if (campaign.closedAt) {
            return <HeroPage icon={AlertCircleIcon} title="Cannot update a closed campaign" />
        }
    }

    const specifyingBranchAllowed =
        // on campaign creation
        (!campaign && patchSet) ||
        // or when it's not yet published and no changesets have been published or are being published as well
        (campaign &&
            !campaign.publishedAt &&
            campaign.changesets.totalCount === 0 &&
            campaign.status.state !== GQL.BackgroundProcessState.PROCESSING)

    const onDraft: React.FormEventHandler = async event => {
        event.preventDefault()
        setMode('saving')
        try {
            const createdCampaign = await createCampaign({
                name,
                description,
                namespace: authenticatedUser.id,
                patchSet: patchSet ? patchSet.id : undefined,
                branch: specifyingBranchAllowed ? branch : undefined,
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
            setAlertError(undefined)
            campaignUpdates.next()
        } catch (err) {
            setAlertError(asError(err))
        } finally {
            setMode('viewing')
        }
    }

    const onSubmit: React.FormEventHandler = async event => {
        event.preventDefault()
        setMode('saving')
        try {
            if (campaignID) {
                const newCampaign = await updateCampaign({
                    id: campaignID,
                    name,
                    description,
                    patchSet: patchSetID ?? undefined,
                    branch: specifyingBranchAllowed ? branch : undefined,
                })
                setCampaign(newCampaign)
                setName(newCampaign.name)
                setDescription(newCampaign.description)
                setBranch(newCampaign.branch ?? '')
                setBranchModified(false)
                unblockHistoryRef.current()
                history.push(`/campaigns/${newCampaign.id}`)
            } else {
                const createdCampaign = await createCampaign({
                    name,
                    description,
                    namespace: authenticatedUser.id,
                    patchSet: patchSet ? patchSet.id : undefined,
                    branch: specifyingBranchAllowed ? branch : undefined,
                })
                unblockHistoryRef.current()
                history.push(`/campaigns/${createdCampaign.id}`)
            }
            setMode('viewing')
            setAlertError(undefined)
            campaignUpdates.next()
        } catch (err) {
            setMode('editing')
            setAlertError(asError(err))
        }
    }

    const discardChangesMessage = 'Do you want to discard your changes?'

    const onEdit: React.MouseEventHandler = event => {
        event.preventDefault()
        unblockHistoryRef.current = history.block(discardChangesMessage)
        setMode('editing')
        setAlertError(undefined)
    }

    const onCancel: React.FormEventHandler = event => {
        event.preventDefault()
        if (!confirm(discardChangesMessage)) {
            return
        }
        unblockHistoryRef.current()
        // clear query params
        history.replace(location.pathname)
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
            setMode('viewing')
        }
    }

    const onRetry = async (): Promise<void> => {
        try {
            setCampaign(await retryCampaign(campaign!.id))
            changesetUpdates.next()
        } catch (err) {
            setAlertError(asError(err))
        }
    }

    const onAddChangeset = (): void => {
        // we also check the campaign.changesets.totalCount, so an update to the campaign is required as well
        campaignUpdates.next()
        changesetUpdates.next()
    }

    const author = campaign ? campaign.author : authenticatedUser

    const onNameChange = (newName: string): void => {
        if (!branchModified) {
            setBranch(slugify(newName, { lower: true }))
        }
        setName(newName)
    }

    const onBranchChange = (event: ChangeEvent<HTMLInputElement>): void => {
        setBranch(event.target.value)
        setBranchModified(true)
    }

    const totalChangesetCount = campaign?.changesets.totalCount ?? 0

    const totalPatchCount = (campaign?.patches.totalCount ?? 0) + (patchSet?.patches.totalCount ?? 0)

    const campaignFormID = 'campaign-form'

    return (
        <>
            <PageTitle title={campaign ? campaign.name : 'New campaign'} />
            <CampaignActionsBar
                previewingPatchSet={!!patchSet}
                mode={mode}
                campaign={campaign}
                onEdit={onEdit}
                onClose={onClose}
                onDelete={onDelete}
                formID={campaignFormID}
            />
            {alertError && <ErrorAlert error={alertError} />}
            <Form id={campaignFormID} onSubmit={onSubmit} onReset={onCancel} className="e2e-campaign-form">
                {campaign && !updateMode && !['saving', 'editing'].includes(mode) && (
                    <CampaignStatus campaign={campaign} onPublish={onPublish} onRetry={onRetry} />
                )}
                {(mode === 'editing' || mode === 'saving') && (
                    <>
                        <h3>Details</h3>
                        <CampaignTitleField
                            className="e2e-campaign-title"
                            value={name}
                            onChange={onNameChange}
                            disabled={mode === 'saving'}
                        />
                        <CampaignDescriptionField
                            value={description}
                            onChange={setDescription}
                            disabled={mode === 'saving'}
                        />
                        {specifyingBranchAllowed && (
                            <div className="form-group mt-2">
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
                                    onChange={onBranchChange}
                                    placeholder="my-awesome-campaign"
                                    value={branch}
                                    required={true}
                                    disabled={mode === 'saving'}
                                />
                            </div>
                        )}
                    </>
                )}
                {/* If we are in the update mode */}
                {updateMode && (
                    <>
                        <CampaignUpdateDiff
                            campaign={campaign!}
                            patchSet={patchSet!}
                            history={history}
                            location={location}
                            isLightTheme={isLightTheme}
                            className="mt-4"
                        />
                        <div className="mb-0">
                            <button
                                type="reset"
                                form={campaignFormID}
                                className="btn btn-secondary mr-1"
                                onClick={onCancel}
                                disabled={mode !== 'editing'}
                            >
                                Cancel
                            </button>
                            <button
                                type="submit"
                                form={campaignFormID}
                                className="btn btn-primary"
                                disabled={mode !== 'editing' || patchSet?.patches.totalCount === 0}
                            >
                                Update
                            </button>
                        </div>
                    </>
                )}
                {/* If campaign doesn't yet exist.. */}
                {!campaign && (
                    <>
                        <div className="mt-2">
                            {/* When creating from a patch set, allow draft campaigns */}
                            {patchSet && (
                                <button
                                    type="submit"
                                    form={campaignFormID}
                                    className="btn btn-secondary mr-1"
                                    // todo: doesn't trigger form validation
                                    onClick={onDraft}
                                    disabled={mode !== 'editing'}
                                >
                                    Create draft
                                </button>
                            )}
                            <button
                                type="submit"
                                form={campaignFormID}
                                className="btn btn-primary"
                                disabled={mode !== 'editing' || patchSet?.patches.totalCount === 0}
                            >
                                Create
                            </button>
                        </div>
                    </>
                )}
            </Form>

            {!updateMode && (campaign || patchSet) && (
                <>
                    {campaign && mode !== 'editing' && mode !== 'saving' && (
                        <div className="card mt-2">
                            <div className="card-header">
                                <strong>
                                    <UserAvatar user={author} className="icon-inline" /> {author.username}
                                </strong>{' '}
                                started <Timestamp date={campaign.createdAt} />
                            </div>
                            <div className="card-body">
                                <Markdown
                                    dangerousInnerHTML={renderMarkdown(campaign.description || '_No description_')}
                                />
                            </div>
                        </div>
                    )}
                    {campaign && !['saving', 'editing'].includes(mode) && (
                        <>
                            <h3 className="mt-4 mb-2">Progress</h3>
                            <CampaignBurndownChart
                                changesetCountsOverTime={campaign.changesetCountsOverTime}
                                history={history}
                            />
                            {/* only campaigns that have no patch set can add changesets manually */}
                            {!campaign.patchSet && campaign.viewerCanAdminister && !campaign.closedAt && (
                                <AddChangesetForm campaignID={campaign.id} onAdd={onAddChangeset} />
                            )}
                        </>
                    )}

                    <h3 className="mt-4 d-flex align-items-end mb-0">
                        {totalPatchCount > 0 && (
                            <>
                                {totalPatchCount} {pluralize('Patch', totalPatchCount, 'Patches')}
                            </>
                        )}
                        {(totalChangesetCount > 0 || !!campaign) && totalPatchCount > 0 && (
                            <span className="mx-1">/</span>
                        )}
                        {(totalChangesetCount > 0 || !!campaign) && (
                            <>
                                {totalChangesetCount} {pluralize('Changeset', totalChangesetCount)}
                            </>
                        )}{' '}
                        {(patchSet || campaign) && (
                            <CampaignDiffStat campaign={campaign} patchSet={patchSet} className="ml-2 mb-0" />
                        )}
                    </h3>
                    {totalChangesetCount + totalPatchCount > 0 ? (
                        <>
                            {totalPatchCount > 0 &&
                                (campaign ? (
                                    <CampaignPatches
                                        campaign={campaign}
                                        campaignUpdates={campaignUpdates}
                                        changesetUpdates={changesetUpdates}
                                        enablePublishing={!campaign.closedAt}
                                        history={history}
                                        location={location}
                                        isLightTheme={isLightTheme}
                                    />
                                ) : (
                                    <PatchSetPatches
                                        patchSet={patchSet!}
                                        campaignUpdates={campaignUpdates}
                                        changesetUpdates={changesetUpdates}
                                        enablePublishing={false}
                                        history={history}
                                        location={location}
                                        isLightTheme={isLightTheme}
                                    />
                                ))}
                            {totalChangesetCount > 0 && (
                                <CampaignChangesets
                                    campaign={campaign!}
                                    changesetUpdates={changesetUpdates}
                                    campaignUpdates={campaignUpdates}
                                    history={history}
                                    location={location}
                                    isLightTheme={isLightTheme}
                                    extensionsController={extensionsController}
                                    platformContext={platformContext}
                                    telemetryService={telemetryService}
                                />
                            )}
                        </>
                    ) : (
                        campaign?.status.state !== GQL.BackgroundProcessState.PROCESSING &&
                        // Show hint for empty manual campaigns
                        (campaign && !campaign.patchSet ? (
                            <div className="mt-2 alert alert-info">Add a changeset to get started.</div>
                        ) : (
                            <p className="mt-2 text-muted">No changesets</p>
                        ))
                    )}
                </>
            )}
        </>
    )
}
