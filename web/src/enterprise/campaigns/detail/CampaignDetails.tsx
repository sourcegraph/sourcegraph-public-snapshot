import { parse as parseJSONC } from '@sqs/jsonc-parser'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { UserAvatar } from '../../../user/UserAvatar'
import { Timestamp } from '../../../components/time/Timestamp'
import { CampaignsIcon } from '../icons'
import { ChangesetNode, ChangesetNodeProps } from './changesets/ChangesetNode'
import { isEqual, noop } from 'lodash'
import { Form } from '../../../components/Form'
import {
    fetchCampaignById,
    updateCampaign,
    deleteCampaign,
    createCampaign,
    queryChangesets,
    previewCampaignPlan,
    fetchCampaignPlanById,
    CampaignType,
    retryCampaign,
} from './backend'
import { useError, useObservable } from '../../../util/useObservable'
import { asError } from '../../../../../shared/src/util/errors'
import * as H from 'history'
import { queryNamespaces } from '../../namespaces/backend'
import { CampaignBurndownChart } from './BurndownChart'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'
import { AddChangesetForm } from './AddChangesetForm'
import { Subject, of, timer, merge, Observable } from 'rxjs'
import { MonacoSettingsEditor } from '../../../settings/MonacoSettingsEditor'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { ErrorAlert } from '../../../components/alerts'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { Link } from '../../../../../shared/src/components/Link'
import { switchMap, tap, catchError, takeWhile, concatMap, repeatWhen, delay } from 'rxjs/operators'
import { ThemeProps } from '../../../../../shared/src/theme'
import { TabsWithLocalStorageViewStatePersistence } from '../../../../../shared/src/components/Tabs'
import { isDefined } from '../../../../../shared/src/util/types'
import { FileDiffTab } from './FileDiffTab'
import combyJsonSchema from '../../../../../schema/campaign-types/comby.schema.json'
import credentialsJsonSchema from '../../../../../schema/campaign-types/credentials.schema.json'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'

interface Props extends ThemeProps {
    /**
     * The campaign ID.
     * If not given, will display a creation form.
     */
    campaignID?: GQL.ID
    authenticatedUser: GQL.IUser
    history: H.History
    location: H.Location
    isSourcegraphDotCom: boolean
}

const jsonSchemaByType: { [K in CampaignType]: any } = {
    comby: combyJsonSchema,
    credentials: credentialsJsonSchema,
}

const defaultInputByType: { [K in CampaignType]: any } = {
    comby: `{
    "scopeQuery": "repo:github.com/foo/bar",
    "matchTemplate": "",
    "rewriteTemplate": ""
}`,
    credentials: `{
    "scopeQuery": "repo:github.com/foo/bar",
    "matchers": [{ "type": "npm" }]
}`,
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
    isSourcegraphDotCom,
}) => {
    // State for the form in editing mode
    const [name, setName] = useState<string>('')
    const [description, setDescription] = useState<string>('')
    const [type, setType] = useState<CampaignType>()
    const [campaignPlanArguments, setCampaignPlanArguments] = useState<string>('')
    const [namespace, setNamespace] = useState<GQL.ID>()

    const [namespaces, setNamespaces] = useState<GQL.Namespace[]>()
    const getNamespace = useCallback((): GQL.ID | undefined => namespace || namespaces?.[0].id, [namespace, namespaces])

    // For errors during fetching
    const triggerError = useError()

    useEffect(() => {
        if (campaignID) {
            // Namespace cannot be edited
            return
        }
        const subscription = queryNamespaces().subscribe({ next: setNamespaces, error: triggerError })
        return () => subscription.unsubscribe()
    }, [campaignID, triggerError])

    const campaignUpdates = useMemo(() => new Subject<void>(), [])

    const changesetUpdates = useMemo(() => new Subject<void>(), [])
    const nextChangesetUpdate = useCallback(changesetUpdates.next.bind(changesetUpdates), [changesetUpdates])

    // Fetch campaign if ID was given
    const [campaign, setCampaign] = useState<GQL.ICampaign | GQL.ICampaignPlan | null>()
    useEffect(() => {
        if (!campaignID) {
            return
        }
        const subscription = merge(of(undefined), campaignUpdates)
            .pipe(
                switchMap(
                    () =>
                        new Observable<GQL.ICampaign | null>(observer => {
                            let currentCampaign: GQL.ICampaign | null
                            const subscription = fetchCampaignById(campaignID)
                                .pipe(
                                    tap(campaign => {
                                        currentCampaign = campaign
                                    }),
                                    repeatWhen(obs =>
                                        obs.pipe(
                                            // todo(a8n): why does this not unsubscribe when takeWhile is in outer pipe
                                            takeWhile(
                                                () =>
                                                    !!currentCampaign &&
                                                    !!currentCampaign.changesetCreationStatus &&
                                                    currentCampaign.changesetCreationStatus.state === 'PROCESSING'
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
                    setType(fetchedCampaign?.plan?.type as CampaignType | undefined)
                    setCampaignPlanArguments(fetchedCampaign?.plan ? fetchedCampaign.plan.arguments : null)
                    nextChangesetUpdate()
                },
                error: triggerError,
            })
        return () => subscription.unsubscribe()
    }, [campaignID, triggerError, nextChangesetUpdate, campaignUpdates])

    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArgs) => queryChangesets(campaignID!, args),
        [campaignID]
    )

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

    const previewCampaignPlans = useMemo(() => new Subject<GQL.ICampaignPlanSpecification>(), [])
    const nextPreviewCampaignPlan = useCallback(previewCampaignPlans.next.bind(previewCampaignPlans), [
        previewCampaignPlans,
    ])
    const [isLoadingPreview, setIsLoadingPreview] = useState<boolean>(false)
    useObservable(
        useMemo(
            () =>
                previewCampaignPlans.pipe(
                    tap(() => {
                        setAlertError(undefined)
                        setIsLoadingPreview(true)
                        setCampaign(undefined)
                    }),
                    switchMap(plan =>
                        previewCampaignPlan(plan, false).pipe(
                            tap(() => {
                                setIsLoadingPreview(false)
                            }),
                            catchError(error => {
                                setAlertError(asError(error))
                                setIsLoadingPreview(false)
                                return []
                            }),
                            switchMap(previewPlan =>
                                merge(
                                    of(previewPlan),
                                    timer(0, 2000).pipe(
                                        concatMap(() => fetchCampaignPlanById(previewPlan.id)),
                                        takeWhile(isDefined),
                                        takeWhile(plan => plan.status.state === 'PROCESSING', true)
                                    )
                                )
                            )
                        )
                    ),
                    tap(setCampaign)
                ),
            [previewCampaignPlans]
        )
    )

    if (campaign === undefined && campaignID) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (campaign === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }

    const onSubmit: React.FormEventHandler = async event => {
        event.preventDefault()
        setMode('saving')
        try {
            if (campaignID) {
                setCampaign(await updateCampaign({ id: campaignID, name, description }))
                unblockHistoryRef.current()
            } else {
                const createdCampaign = await createCampaign({
                    name,
                    description,
                    namespace: getNamespace()!,
                    plan: campaign && campaign.__typename === 'CampaignPlan' ? campaign.id : undefined,
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

    const onChangeArguments = (newText: string): void => {
        setCampaignPlanArguments(newText)
    }

    const onChangeType = (event: React.ChangeEvent<HTMLSelectElement>): void => {
        const newType = (event.target.value as CampaignType) || undefined
        const parsedContent = parseJSONC(campaignPlanArguments)
        if ((newType && !parsedContent) || (type && isEqual(parsedContent, parseJSONC(defaultInputByType[type])))) {
            setCampaignPlanArguments(defaultInputByType[newType])
        }
        setType(newType)
        setCampaign(undefined)
    }

    const discardChangesMessage = 'Do you want to discard your changes?'

    const onEdit: React.MouseEventHandler = event => {
        event.preventDefault()
        unblockHistoryRef.current = history.block(discardChangesMessage)
        {
            const { name, description, plan } = campaign as GQL.ICampaign
            setName(name)
            setDescription(description)
            setMode('editing')
            setType(plan?.type as CampaignType | undefined)
            setCampaignPlanArguments(plan ? plan.arguments : '')
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

    const onRetry: React.MouseEventHandler = async event => {
        event.preventDefault()
        try {
            await retryCampaign(campaign!.id)
            campaignUpdates.next()
        } catch (err) {
            setAlertError(asError(err))
        }
    }

    const author = campaign && campaign.__typename === 'Campaign' ? campaign.author : authenticatedUser

    const nodes: (GQL.IExternalChangeset | GQL.IChangesetPlan)[] | undefined = campaign?.changesets.nodes

    const calculateDiff = (field: 'added' | 'deleted'): number => {
        if (!nodes) {
            return 0
        }
        return nodes.reduce(
            (prev, next) =>
                prev + (next.diff ? next.diff.fileDiffs.diffStat[field] + next.diff.fileDiffs.diffStat.changed : 0),
            0
        )
    }

    const totalAdditions = calculateDiff('added')
    const totalDeletions = calculateDiff('deleted')

    const status = campaign
        ? campaign.__typename === 'CampaignPlan'
            ? campaign.status
            : campaign.changesetCreationStatus
        : null

    const currentSpec = campaign && campaign.__typename === 'CampaignPlan' ? parseJSONC(campaign.arguments) : undefined
    // Tracks if a refresh of the campaignPlan is required before the campaign can be created
    const previewRefreshNeeded =
        !currentSpec ||
        !isEqual(currentSpec, parseJSONC(campaignPlanArguments)) ||
        (status && status.state !== 'COMPLETED')

    return (
        <>
            <PageTitle title={campaign && campaign.__typename === 'Campaign' ? campaign.name : 'New Campaign'} />
            <Form onSubmit={onSubmit} onReset={onCancel} className="e2e-campaign-form">
                <h2 className="d-flex">
                    <CampaignsIcon className="icon-inline mr-2" />
                    <span>
                        <Link to="/campaigns">Campaigns</Link>
                    </span>
                    <span className="text-muted d-inline-block mx-2">/</span>
                    {/* The namespace of a campaign can only be set on creation */}
                    {campaign && campaign.__typename === 'Campaign' ? (
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
                            {namespaces?.map(namespace => (
                                <option value={namespace.id} key={namespace.id}>
                                    {namespace.namespaceName}
                                </option>
                            ))}
                        </select>
                    )}
                    <span className="text-muted d-inline-block mx-2">/</span>
                    {mode === 'editing' || mode === 'saving' ? (
                        <input
                            className="form-control w-auto d-inline-block e2e-campaign-title"
                            value={name}
                            onChange={event => setName(event.target.value)}
                            placeholder="Campaign title"
                            disabled={mode === 'saving'}
                            autoFocus={true}
                            required={true}
                        />
                    ) : (
                        <span>{campaign && campaign.__typename === 'Campaign' && campaign.name}</span>
                    )}
                    <span className="flex-grow-1 d-flex justify-content-end align-items-center">
                        {(mode === 'saving' || mode === 'deleting') && <LoadingSpinner className="mr-2" />}
                        {campaignID &&
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
                            ))}
                    </span>
                </h2>
                {alertError && <ErrorAlert error={alertError} />}
                <div className="card">
                    <div className="card-header">
                        <strong>
                            <UserAvatar user={author} className="icon-inline" /> {author.username}
                        </strong>
                        {campaign && campaign.__typename === 'Campaign' && (
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
                            rows={8}
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
                            <a
                                rel="noopener noreferrer"
                                target="_blank"
                                href={
                                    (isSourcegraphDotCom ? 'https://docs.sourcegraph.com' : '/help') + '/user/markdown'
                                }
                            >
                                Markdown supported
                            </a>
                        </small>
                    </p>
                )}
                <h3 className="mt-3">Campaign type</h3>
                <select
                    className="form-control w-auto d-inline-block e2e-campaign-type"
                    placeholder="Select campaign type"
                    onChange={onChangeType}
                    disabled={!!(campaign && campaign.__typename === 'Campaign')}
                    value={type}
                >
                    <option value="">Manual</option>
                    <option value="comby">Comby search and replace</option>
                    <option value="credentials">NPM Credentials</option>
                </select>
                {type === 'comby' && (
                    <small className="ml-1">
                        <a rel="noopener noreferrer" target="_blank" href="https://comby.dev/#match-syntax">
                            Learn about comby syntax
                        </a>
                    </small>
                )}
                <MonacoSettingsEditor
                    className="my-3 e2e-campaign-arguments"
                    isLightTheme={isLightTheme}
                    value={campaignPlanArguments}
                    jsonSchema={type ? jsonSchemaByType[type] : undefined}
                    height={110}
                    onChange={onChangeArguments}
                    readOnly={!!(campaign && campaign.__typename === 'Campaign')}
                ></MonacoSettingsEditor>
                {(!campaign || (campaign && campaign.__typename === 'CampaignPlan')) && mode === 'editing' && (
                    <>
                        {type !== undefined && (
                            <button
                                type="button"
                                className="btn btn-primary mr-1 e2e-preview-campaign"
                                disabled={!previewRefreshNeeded}
                                onClick={() => nextPreviewCampaignPlan({ type, arguments: campaignPlanArguments })}
                            >
                                {isLoadingPreview && <LoadingSpinner className="icon-inline mr-1" />}
                                Preview changes
                            </button>
                        )}
                        <button
                            type="submit"
                            className="btn btn-primary"
                            disabled={(type !== undefined && previewRefreshNeeded) || mode !== 'editing'}
                        >
                            Create
                        </button>
                    </>
                )}
            </Form>

            {status && (
                <>
                    {status.state === 'PROCESSING' && (
                        <div className="d-flex mt-3 e2e-preview-loading">
                            <LoadingSpinner className="icon-inline" />{' '}
                            <span data-tooltip="Computing changesets">
                                {status.completedCount} / {status.pendingCount + status.completedCount}
                            </span>
                        </div>
                    )}
                    {type && status.state !== 'PROCESSING' && (
                        <div className="d-flex my-3">
                            {status.state === 'COMPLETED' && (
                                <CheckCircleIcon className="icon-inline text-success mr-1 e2e-preview-success" />
                            )}
                            {status.state === 'ERRORED' && <AlertCircleIcon className="icon-inline text-danger mr-1" />}{' '}
                            {/* Status asserts on campaign being set, this will never be null */}
                            {campaign!.__typename === 'Campaign' ? 'Creation' : 'Preview'}{' '}
                            {status.state.toLocaleLowerCase()}
                        </div>
                    )}
                    {status.errors.map((error, i) => (
                        <ErrorAlert error={error} className="mt-3" key={i} />
                    ))}
                    {status.state === 'ERRORED' && campaign?.__typename === 'Campaign' && (
                        <button type="button" className="btn btn-primary mb-2" onClick={onRetry}>
                            Retry failed jobs
                        </button>
                    )}
                </>
            )}

            {campaign && campaign.__typename === 'Campaign' && (
                <>
                    <h3>Progress</h3>
                    <CampaignBurndownChart
                        changesetCountsOverTime={campaign.changesetCountsOverTime}
                        history={history}
                    />
                    {/* only campaigns that have no plan can add changesets manually */}
                    {!campaign.plan && <AddChangesetForm campaignID={campaign.id} onAdd={nextChangesetUpdate} />}
                </>
            )}
            {/* is already created or a preview is available */}
            {campaign && (
                <>
                    <TabsWithLocalStorageViewStatePersistence
                        storageKey="campaignTab"
                        className="mt-3"
                        tabs={[
                            {
                                id: 'diff',
                                label: (
                                    <span>
                                        Diff <span className="text-success">+{totalAdditions}</span>{' '}
                                        <span className="text-danger">-{totalDeletions}</span>
                                    </span>
                                ),
                            },
                            {
                                id: 'changesets',
                                label: (
                                    <span>
                                        Changesets{' '}
                                        {campaign && (
                                            <span className="badge badge-secondary badge-pill">
                                                {campaign.changesets.totalCount}
                                            </span>
                                        )}
                                    </span>
                                ),
                            },
                        ]}
                        tabClassName="tab-bar__tab--h5like"
                    >
                        <div className="list-group mt-3" key="changesets">
                            {campaign && campaign.__typename === 'Campaign' && (
                                <FilteredConnection<
                                    GQL.IExternalChangeset | GQL.IChangesetPlan,
                                    Omit<ChangesetNodeProps, 'node'>
                                >
                                    className="mt-2"
                                    updates={changesetUpdates}
                                    nodeComponent={ChangesetNode}
                                    nodeComponentProps={{ isLightTheme, history, location }}
                                    queryConnection={queryChangesetsConnection}
                                    hideSearch={true}
                                    defaultFirst={15}
                                    noun="Changeset"
                                    pluralNoun="Changesets"
                                    history={history}
                                    location={location}
                                />
                            )}
                            {campaign &&
                                campaign.__typename === 'CampaignPlan' &&
                                campaign.changesets.nodes.map((changeset, i) => (
                                    <ChangesetNode
                                        node={changeset}
                                        isLightTheme={isLightTheme}
                                        key={i}
                                        location={location}
                                        history={history}
                                    ></ChangesetNode>
                                ))}
                        </div>
                        <div className="mt-3" key="diff">
                            {nodes && (
                                <FileDiffTab
                                    nodes={nodes}
                                    persistLines={campaign.__typename === 'Campaign'}
                                    history={history}
                                    location={location}
                                    isLightTheme={isLightTheme}
                                ></FileDiffTab>
                            )}
                        </div>
                    </TabsWithLocalStorageViewStatePersistence>
                </>
            )}
        </>
    )
}
