import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignsIcon } from '../icons'

interface Props {
    node: GQL.ICampaign
}

/**
 * An item in the list of campaigns.
 */
export const CampaignNode: React.FunctionComponent<Props> = ({ node }) => (
    <li className="card p-2 mt-2">
        <div className="d-flex">
            <CampaignsIcon className="icon-inline mr-2" />
            <div>
                <h3 className="mb-0">
                    <Link to={`/campaigns/${node.id}`} className="d-flex align-items-center text-decoration-none">
                        {node.name}
                    </Link>
                </h3>
                <div>{node.description}</div>
            </div>
        </div>
    </li>
)
