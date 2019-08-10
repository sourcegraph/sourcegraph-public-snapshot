import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
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
            <h5 className="card-header">
                Preview
                {isLoading && <LoadingSpinner className="icon-inline ml-2" />}
            </h5>
            {campaignPreview !== LOADING &&
                (isErrorLike(campaignPreview) ? (
                    <div className="alert alert-danger border-0">Error: {campaignPreview.message}</div>
                ) : (
                    <div className="card-body" style={isLoading ? { opacity: '0.7', cursor: 'wait' } : undefined}>
                        <ul>
                            <li>Diagnostics: {campaignPreview.diagnostics.totalCount}</li>
                            <li>
                                Threads: {campaignPreview.threads.totalCount}{' '}
                                {campaignPreview.threads.nodes.map(t => t.title).join(', ')}
                            </li>
                        </ul>
                    </div>
                ))}
        </div>
    )
}
