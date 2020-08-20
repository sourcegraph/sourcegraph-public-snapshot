import * as H from 'history'
import React, { useCallback } from 'react'
import { CampaignSpecFields } from '../../../graphql-operations'
import { createCampaign, applyCampaign } from './backend'
import { Link } from '../../../../../shared/src/components/Link'
import classNames from 'classnames'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'

export interface CreateUpdateCampaignAlertProps {
    specID: string
    campaign: CampaignSpecFields['appliesToCampaign']
    isLoading: boolean | Error
    setIsLoading: (newValue: boolean | Error) => void
    viewerCanAdminister: boolean
    history: H.History
}

export const CreateUpdateCampaignAlert: React.FunctionComponent<CreateUpdateCampaignAlertProps> = ({
    specID,
    campaign,
    isLoading,
    setIsLoading,
    viewerCanAdminister,
    history,
}) => {
    const campaignID = campaign?.id
    const onApply = useCallback(async () => {
        if (!confirm(`Are you sure you want to ${campaignID ? 'update' : 'create'} this campaign?`)) {
            return
        }
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
            <div className="alert alert-info p-3 mb-3 d-flex align-items-center body-lead">
                <h2 className="m-0 mr-3">
                    <span className="badge badge-info text-uppercase mb-0">Preview</span>
                </h2>
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
                        'btn btn-primary test-campaigns-confirm-apply-btn',
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
