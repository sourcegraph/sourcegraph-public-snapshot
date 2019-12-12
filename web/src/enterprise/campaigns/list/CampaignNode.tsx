import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { CampaignsIcon } from '../icons'
import { Link } from '../../../../../shared/src/components/Link'

interface Props {
    node: GQL.ICampaign
}

/**
 * An item in the list of campaigns.
 */
export const CampaignNode: React.FunctionComponent<Props> = ({ node }) => (
    <li className="card p-2 mt-2">
        <div className="d-flex">
            <CampaignsIcon className="icon-inline mr-2 flex-shrink-0" />
            <div className="campaign-node__content">
                <h3 className="m-0">
                    <Link to={`/campaigns/${node.id}`} className="d-flex align-items-center text-decoration-none">
                        {node.name}
                    </Link>
                </h3>
                <Markdown
                    className="text-truncate"
                    dangerousInnerHTML={renderMarkdown(node.description, { plainText: true })}
                ></Markdown>
            </div>
        </div>
    </li>
)
