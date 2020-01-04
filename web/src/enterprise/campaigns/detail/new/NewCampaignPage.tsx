import React, { useState, useEffect, useRef, useCallback } from 'react'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import classNames from 'classnames'
import H from 'history'
import { noop, isEqual } from 'lodash'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { NewCampaignFormData, NewCampaignForm, isExistingPlanID } from './NewCampaignForm'
import {
    DEFAULT_CAMPAIGN_PLAN_SPECIFICATION_FORM_DATA,
    MANUAL_CAMPAIGN_TYPE,
} from '../form/CampaignPlanSpecificationFields'
import { PageTitle } from '../../../../components/PageTitle'
import { Link } from '../../../../../../shared/src/components/Link'
import { Form } from '../../../../components/Form'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { createCampaign } from '../backend'
import { asError, isErrorLike } from '../../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../../components/alerts'
import { CampaignTabs } from '../CampaignTabs'
import { useCampaignPlan } from './useCampaignPlan'

interface Props extends ThemeProps {
    authenticatedUser: Pick<GQL.IUser, 'id' | 'username' | 'avatarURL'>

    history: H.History
    location: H.Location

    /** For testing only. */
    _useCampaignPlan?: typeof useCampaignPlan
}

export const NewCampaignPage: React.FunctionComponent<Props> = ({
    authenticatedUser,
    history,
    location,
    isLightTheme,
    _useCampaignPlan = useCampaignPlan,
}) => {
    const planID: GQL.ID | null = new URLSearchParams(location.search).get('plan')

    const [value, setValue] = useState<NewCampaignFormData>({
        name: '',
        description: '',
        plan: planID || DEFAULT_CAMPAIGN_PLAN_SPECIFICATION_FORM_DATA,
    })

    const [mode, setMode] = useState<'editing' | 'saving' | 'deleting' | 'closing'>('editing')

    // To report errors from saving or deleting.
    const [alertError, setAlertError] = useState<Error>()

    // To unblock the history after leaving edit mode.
    const unblockHistoryRef = useRef<H.UnregisterCallback>(noop)
    useEffect(() => {
        unblockHistoryRef.current()
        unblockHistoryRef.current = history.block('Do you want to discard this campaign?')
        return unblockHistoryRef.current
    }, [history])

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setMode('saving')
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            ;(async () => {
                try {
                    const createdCampaign = await createCampaign({
                        name: value.name,
                        description: value.description,
                        namespace: authenticatedUser.id,
                        plan: undefined, // TODO!(sqs)
                    })
                    unblockHistoryRef.current()
                    history.push(`/campaigns/${createdCampaign.id}`)
                    setAlertError(undefined)
                } catch (err) {
                    setMode('editing')
                    setAlertError(asError(err))
                }
            })()
        },
        [authenticatedUser.id, history, value.description, value.name]
    )

    const [committedCampaignPlanSpecOrID, setCommittedCampaignPlanSpecOrID] = useState<
        GQL.ICampaignPlanSpecification | GQL.ID | undefined
    >(typeof value.plan === 'string' ? value.plan : undefined)
    const [campaignPlan, isLoadingCampaignPlan] = _useCampaignPlan(committedCampaignPlanSpecOrID)
    const showPreview = Boolean(!isExistingPlanID(value.plan) && value.plan.type !== MANUAL_CAMPAIGN_TYPE)
    const isPlanStale = !isEqual(value.plan, committedCampaignPlanSpecOrID)

    // Clear the fetched campaign plan when significant changes are made to the spec.
    const campaignPlanSpecType = !isExistingPlanID(value.plan) ? value.plan.type : undefined
    useEffect(() => {
        setCommittedCampaignPlanSpecOrID(undefined)
    }, [campaignPlanSpecType])

    const onPreviewClick = useCallback((): void => setCommittedCampaignPlanSpecOrID(value.plan), [value.plan])

    return (
        <>
            <PageTitle title="New campaign" />
            <nav className="mb-2" aria-label="breadcrumb">
                <ol className="breadcrumb">
                    <li className="breadcrumb-item">
                        <Link to="/campaigns" className="text-decoration-none">
                            Campaigns
                        </Link>
                    </li>
                    <li className="breadcrumb-item active" aria-current="page">
                        New
                    </li>
                </ol>
            </nav>
            <Form onSubmit={onSubmit} className="e2e-campaign-form">
                <NewCampaignForm
                    value={value}
                    onChange={setValue}
                    disabled={mode === 'saving'}
                    isLightTheme={isLightTheme}
                >
                    <div className="card-body p-3 d-flex align-items-center">
                        {showPreview && (
                            <button
                                type="button"
                                className={classNames(
                                    'btn mr-1 e2e-preview-campaign',
                                    isPlanStale ? 'btn-primary' : 'btn-secondary'
                                )}
                                disabled={!isPlanStale || mode === 'saving'}
                                onClick={onPreviewClick}
                            >
                                Preview changes
                            </button>
                        )}
                        <button
                            type="submit"
                            className={classNames(
                                'btn e2e-create-campaign',
                                isPlanStale ? 'btn-secondary' : 'btn-success'
                            )}
                            disabled={isPlanStale || mode === 'saving'}
                        >
                            Create
                        </button>
                        {isLoadingCampaignPlan && <LoadingSpinner className="icon-inline ml-2" />}
                    </div>
                    {alertError && <ErrorAlert error={alertError} />}
                </NewCampaignForm>
            </Form>
            {campaignPlan && !isErrorLike(campaignPlan) && (
                <CampaignTabs
                    changesets={campaignPlan.changesets}
                    persistLines={false}
                    history={history}
                    location={location}
                    className="mt-3"
                    isLightTheme={isLightTheme}
                />
            )}
        </>
    )
}
