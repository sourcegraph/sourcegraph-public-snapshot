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

export interface CampaignNodeProps {
    node: Pick<
        GQL.ICampaign,
        'id' | 'closedAt' | 'name' | 'description' | 'changesets' | 'changesetPlans' | 'createdAt'
    >
    /** Renders a selection button next to the campaign, used to select a campaign for update */
    selection?: {
        enabled: boolean
        buttonLabel: string
        onSelect: (campaign: Pick<GQL.ICampaign, 'id'>) => void
    }
    /** Used for testing purposes. Sets the current date */
    now?: Date
}

/**
 * An item in the list of campaigns.
 */
export const CampaignNode: React.FunctionComponent<CampaignNodeProps> = ({ node, selection, now = new Date() }) => {
    const campaignIconClass = node.closedAt ? 'text-danger' : 'text-success'
    const OpenChangesetIcon = changesetStateIcons[GQL.ChangesetState.OPEN]
    const MergedChangesetIcon = changesetStateIcons[GQL.ChangesetState.MERGED]
    const changesetCountByState = (state: GQL.ChangesetState): number =>
        node.changesets.nodes.reduce((prev, next) => prev + (next.state === state ? 1 : 0), 0)
    return (
        <li className="card p-2 mt-2">
            <div className="d-flex align-items-center">
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
                        className={classNames(
                            'text-truncate',
                            !node.description && 'text-muted',
                            !node.description && 'font-italic'
                        )}
                        dangerousInnerHTML={
                            node.description ? renderMarkdown(node.description, { plainText: true }) : 'No description'
                        }
                    ></Markdown>
                </div>
                <div data-tooltip="Open changesets">
                    {changesetCountByState(GQL.ChangesetState.OPEN) + node.changesetPlans.totalCount}{' '}
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
