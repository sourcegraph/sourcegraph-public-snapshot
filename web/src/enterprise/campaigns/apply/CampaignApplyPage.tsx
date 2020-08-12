import * as H from 'history'
import React, { useMemo, useCallback, useState } from 'react'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { PageTitle } from '../../../components/PageTitle'
import { fetchCampaignSpecById, createCampaign, applyCampaign } from './backend'
import { ErrorAlert } from '../../../components/alerts'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { CampaignHeader } from '../detail/CampaignHeader'
import { ChangesetSpecList } from './ChangesetSpecList'
import { ThemeProps } from '../../../../../shared/src/theme'
import { Link } from '../../../../../shared/src/components/Link'
import classNames from 'classnames'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { CampaignSpecFields } from '../../../graphql-operations'
import { Timestamp } from '../../../components/time/Timestamp'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'

export interface CampaignApplyPageProps extends ThemeProps {
    specID: string
    history: H.History
    location: H.Location
}

export const CampaignApplyPage: React.FunctionComponent<CampaignApplyPageProps> = ({
    specID,
    history,
    location,
    isLightTheme,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const spec = useObservable(useMemo(() => fetchCampaignSpecById(specID), [specID]))
    if (spec === undefined) {
        return <LoadingSpinner />
    }
    if (spec === null) {
        return <ErrorAlert history={history} error={new Error('Campaign spec not found')} />
    }
    return (
        <>
            <PageTitle title="Apply campaign spec" />
            <div className="mb-3">
                <CampaignHeader name={spec.description.name} namespace={spec.namespace} className="d-inline-block" />
                <span className="text-muted ml-3">
                    Uploaded <Timestamp date={spec.createdAt} /> by{' '}
                    {spec.creator && <Link to={spec.creator.url}>{spec.creator.username}</Link>}
                    {!spec.creator && <strong>deleted user</strong>}
                </span>
            </div>
            <CreateUpdateCampaignAlert
                history={history}
                specID={spec.id}
                campaign={spec.appliesToCampaign}
                isLoading={isLoading}
                setIsLoading={setIsLoading}
                viewerCanAdminister={spec.viewerCanAdminister}
            />
            <h2 className="mb-3">What this does</h2>
            <Markdown
                dangerousInnerHTML={renderMarkdown(spec.description.description || '_No description_')}
                history={history}
                className="mb-3"
            />
            <ChangesetSpecList
                campaignSpecID={specID}
                history={history}
                location={location}
                isLightTheme={isLightTheme}
            />
            <CreateUpdateCampaignAlert
                history={history}
                specID={spec.id}
                campaign={spec.appliesToCampaign}
                isLoading={isLoading}
                setIsLoading={setIsLoading}
                viewerCanAdminister={spec.viewerCanAdminister}
            />
        </>
    )
}

export const CreateUpdateCampaignAlert: React.FunctionComponent<{
    specID: string
    campaign: CampaignSpecFields['appliesToCampaign']
    isLoading: boolean | Error
    setIsLoading: (newValue: boolean | Error) => void
    viewerCanAdminister: boolean
    history: H.History
}> = ({ specID, campaign, isLoading, setIsLoading, viewerCanAdminister, history }) => {
    const campaignID = campaign?.id
    const onApply = useCallback(async () => {
        setIsLoading(true)
        try {
            const campaign = campaignID
                ? await applyCampaign({ campaignSpec: specID, campaign: campaignID })
                : await createCampaign({ campaignSpec: specID })
            history.push(campaign.url)
            setIsLoading(false)
        } catch (error) {
            setIsLoading(error)
        }
    }, [specID, setIsLoading, history, campaignID])
    return (
        <>
            <div className="alert alert-info p-3 mb-3 d-flex align-items-center">
                <span className="badge badge-info text-uppercase mb-0 mr-3">Preview</span>
                {!campaign && (
                    <p className="mb-0 flex-grow-1">
                        This campaign is in preview mode. Click create campaign to publish it.
                    </p>
                )}
                {campaign && (
                    <p className="mb-0 flex-grow-1">
                        This operation will update the existing campaign <Link to={campaign.url}>{campaign.name}</Link>.
                        Click update campaign to accept the changes.
                    </p>
                )}
                <button
                    type="button"
                    className={classNames(
                        'btn btn-primary',
                        isLoading === true || (!viewerCanAdminister && 'disabled')
                    )}
                    onClick={onApply}
                    disabled={isLoading === true || !viewerCanAdminister}
                    data-tooltip={!viewerCanAdminister ? 'You have no permission to apply this campaign.' : undefined}
                >
                    {campaignID ? 'Update' : 'Create'} campaign
                </button>
            </div>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} history={history} />}
        </>
    )
}
