import React from 'react'
import { CampaignFormData } from '../CampaignForm'

interface Props {
    data: CampaignFormData | undefined

    className?: string
}

/**
 * A campaign preview.
 */
export const CampaignPreview: React.FunctionComponent<Props> = ({ data, className = '' }) => (
    <div className={`campaign-preview border ${className}`}>${JSON.stringify(data)}</div>
)
