import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { isErrorLike } from '../../../../../../../shared/src/util/errors'
import { CampaignFormData } from '../CampaignForm'
import { useCampaignPreview } from './useCampaignPreview'

interface Props {
    data: CampaignFormData

    className?: string
}

const LOADING = 'loading' as const

/**
 * A campaign preview.
 */
export const CampaignPreview: React.FunctionComponent<Props> = ({ data, className = '' }) => {
    const campaignPreview = useCampaignPreview(data)
    return (
        <div className={`campaign-preview ${className}`}>
            {campaignPreview === LOADING ? (
                <LoadingSpinner />
            ) : isErrorLike(campaignPreview) ? (
                <div className="alert alert-danger">Error: {campaignPreview.message}</div>
            ) : (
                JSON.stringify(campaignPreview)
            )}
        </div>
    )
}
