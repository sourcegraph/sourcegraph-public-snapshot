import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Link } from '../../../../../shared/src/components/Link'

interface Props {
    campaign: Pick<GQL.ICampaign, 'name' | 'url'> | null
    className?: string
}

export const CampaignBreadcrumbs: React.FunctionComponent<Props> = ({ campaign, className = '' }) => (
    <nav aria-label="breadcrumb" className={className}>
        <ol className="breadcrumb">
            {campaign ? (
                <>
                    <li className="breadcrumb-item">
                        <Link to="/campaigns">Campaigns</Link>
                    </li>
                    <li className="breadcrumb-item active">
                        <Link to={campaign.url}>{campaign.name}</Link>
                    </li>
                </>
            ) : (
                <li className="breadcrumb-item active" aria-current="page">
                    Campaigns
                </li>
            )}
        </ol>
    </nav>
)
