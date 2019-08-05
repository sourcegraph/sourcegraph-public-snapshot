import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { CampaignsIcon } from '../icons'
import { useObjectCampaigns } from './useObjectCampaigns'

interface Props {
    /** The object whose campaigns to list. */
    object: Pick<GQL.CampaignNode, 'id'>

    icon?: boolean

    itemClassName?: string
}

const LOADING = 'loading' as const

/**
 * A list of campaigns that contain an object.
 */
export const ObjectCampaignsList: React.FunctionComponent<Props> = ({ object, icon, itemClassName }) => {
    const [campaigns] = useObjectCampaigns(object)
    return campaigns === LOADING ? (
        <LoadingSpinner className="icon-inline" />
    ) : isErrorLike(campaigns) ? (
        <div className="alert alert-danger">{campaigns.message}</div>
    ) : campaigns.totalCount > 0 ? (
        <ul className="list-unstyled">
            {campaigns.nodes.map(campaign => (
                <li key={campaign.id} className={`text-truncate ${itemClassName}`}>
                    <Link to={campaign.url}>
                        {icon && <CampaignsIcon className="icon-inline small mr-1" />} {campaign.name}
                    </Link>
                </li>
            ))}
        </ul>
    ) : (
        <small className="text-muted">Not in any campaigns</small>
    )
}
