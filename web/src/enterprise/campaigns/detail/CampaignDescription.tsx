import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Timestamp } from '../../../components/time/Timestamp'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import H from 'history'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { Link } from 'react-router-dom'

interface Props {
    campaign: Pick<GQL.ICampaign, 'description' | 'updatedAt'> & {
        author: Pick<GQL.ICampaign['author'], 'username' | 'url'>
    }
    history: H.History
    className?: string
}

export const CampaignDescription: React.FunctionComponent<Props> = ({ campaign, history, className = '' }) => (
    <div className={`card ${className}`}>
        <div className="card-header">
            <Link to={campaign.author.url}>
                <strong>{campaign.author.username}</strong>
            </Link>{' '}
            <span className="text-muted">
                <Timestamp date={campaign.updatedAt} />
            </span>
        </div>
        <div className="card-body">
            <Markdown
                dangerousInnerHTML={renderMarkdown(campaign.description || '_No description_')}
                history={history}
            />
        </div>
    </div>
)
