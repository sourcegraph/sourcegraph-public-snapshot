import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { CampaignsIcon } from '../icons'
import { Link } from '../../../../../shared/src/components/Link'
import classNames from 'classnames'
import { changesetStateIcons, changesetStatusColorClasses } from '../detail/changesets/presentation'
import formatDistance from 'date-fns/formatDistance'
import parseISO from 'date-fns/parseISO'
import * as H from 'history'

export type CampaignNodeCampaign = Pick<GQL.ICampaign, 'id' | 'closedAt' | 'name' | 'description' | 'createdAt'> & {
    changesets: {
        nodes: Pick<GQL.IExternalChangeset, 'state'>[]
    }
}

export interface CampaignNodeProps {
    node: CampaignNodeCampaign
    /** Renders a selection button next to the campaign, used to select a campaign for update */
    selection?: {
        enabled: boolean
        buttonLabel: string
        onSelect: (campaign: Pick<GQL.ICampaign, 'id'>) => void
    }
    /** Used for testing purposes. Sets the current date */
    now?: Date
    history: H.History
}

/**
 * An item in the list of campaigns.
 */
export const CampaignNode: React.FunctionComponent<CampaignNodeProps> = ({
    node,
    selection,
    history,
    now = new Date(),
}) => {
    const campaignIconClass = node.closedAt ? 'text-danger' : 'text-success'
    const OpenChangesetIcon = changesetStateIcons[GQL.ChangesetState.OPEN]
    const MergedChangesetIcon = changesetStateIcons[GQL.ChangesetState.MERGED]
    const changesetCountByState = (state: GQL.ChangesetState): number =>
        node.changesets.nodes.reduce((previous, next) => previous + (next.state === state ? 1 : 0), 0)
    return (
        <li className="list-group-item">
            <div className="d-flex align-items-center p-2">
                <CampaignsIcon
                    className={classNames('icon-inline mr-2 flex-shrink-0 align-self-stretch', campaignIconClass)}
                    data-tooltip={node.closedAt ? 'Closed' : 'Open'}
                />
                <div className="campaign-node__content">
                    <div className="m-0 d-flex align-items-baseline">
                        <h3 className="m-0 d-inline-block">
                            <Link to={`/campaigns/${node.id}`}>{node.name}</Link>
                        </h3>
                        <small className="ml-2 text-muted" data-tooltip={node.createdAt}>
                            created {formatDistance(parseISO(node.createdAt), now)} ago
                        </small>
                    </div>
                    <Markdown
                        className={classNames('text-truncate', !node.description && 'text-muted font-italic')}
                        dangerousInnerHTML={
                            node.description ? renderMarkdown(node.description, { plainText: true }) : 'No description'
                        }
                        history={history}
                    />
                </div>
                <div data-tooltip="Open changesets">
                    {changesetCountByState(GQL.ChangesetState.OPEN)}{' '}
                    <OpenChangesetIcon className={`text-${changesetStatusColorClasses.OPEN} ml-1 mr-2`} />
                </div>
                <div data-tooltip="Merged changesets">
                    {changesetCountByState(GQL.ChangesetState.MERGED)}{' '}
                    <MergedChangesetIcon className={`text-${changesetStatusColorClasses.MERGED} ml-1`} />
                </div>
                {selection?.enabled && (
                    <button type="button" className="btn btn-secondary ml-3" onClick={() => selection.onSelect(node)}>
                        {selection.buttonLabel}
                    </button>
                )}
            </div>
        </li>
    )
}
