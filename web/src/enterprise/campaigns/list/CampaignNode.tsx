import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { CampaignsIcon } from '../icons'
import { Link } from '../../../../../shared/src/components/Link'
import classNames from 'classnames'
import formatDistance from 'date-fns/formatDistance'
import parseISO from 'date-fns/parseISO'
import * as H from 'history'
import { changesetExternalStateIcons, changesetExternalStateColorClasses } from '../detail/changesets/presentation'

export type CampaignNodeCampaign = Pick<GQL.ICampaign, 'id' | 'closedAt' | 'name' | 'description' | 'createdAt'> & {
    changesets: {
        nodes: Pick<GQL.IExternalChangeset, 'externalState'>[]
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
    const OpenChangesetIcon = changesetExternalStateIcons[GQL.ChangesetExternalState.OPEN]
    const MergedChangesetIcon = changesetExternalStateIcons[GQL.ChangesetExternalState.MERGED]
    const changesetCountByState = (state: GQL.ChangesetExternalState): number =>
        node.changesets.nodes.reduce((previous, next) => previous + (next.externalState === state ? 1 : 0), 0)
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
                    {changesetCountByState(GQL.ChangesetExternalState.OPEN)}{' '}
                    <OpenChangesetIcon className={`text-${changesetExternalStateColorClasses.OPEN} ml-1 mr-2`} />
                </div>
                <div data-tooltip="Merged changesets">
                    {changesetCountByState(GQL.ChangesetExternalState.MERGED)}{' '}
                    <MergedChangesetIcon className={`text-${changesetExternalStateColorClasses.MERGED} ml-1`} />
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
