import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import EyeIcon from 'mdi-react/EyeIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../../../shared/src/util/errors'
import { CampaignFormData } from '../CampaignForm'
import { useCampaignPreview } from './useCampaignPreview'

interface Props extends ExtensionsControllerProps {
    data: CampaignFormData

    className?: string
}

const LOADING = 'loading' as const

/**
 * A campaign preview.
 */
export const CampaignPreview: React.FunctionComponent<Props> = ({ data, className = '', extensionsController }) => {
    const [campaignPreview, isLoading] = useCampaignPreview({ extensionsController }, data)
    return (
        <div className={`card campaign-preview ${className}`}>
            <h4 className="card-header">
                <EyeIcon className="icon-inline" /> Preview
                {isLoading && <LoadingSpinner className="icon-inline ml-2" />}
            </h4>
            {campaignPreview !== LOADING &&
                (isErrorLike(campaignPreview) ? (
                    <div className="alert alert-danger border-0">Error: {campaignPreview.message}</div>
                ) : (
                    <div className="card-body" style={isLoading ? { opacity: '0.7', cursor: 'wait' } : undefined}>
                        asdf
                    </div>
                ))}
        </div>
    )
}
