import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { Link } from '../../../../../shared/src/components/Link'
import classNames from 'classnames'
import formatDistance from 'date-fns/formatDistance'
import parseISO from 'date-fns/parseISO'
import * as H from 'history'
import { Timestamp } from '../../../components/time/Timestamp'
import { ListCampaign } from '../../../graphql-operations'
import {
    ChangesetStatusOpen,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
} from '../detail/changesets/ChangesetStatusCell'

export interface CampaignNodeProps {
    node: ListCampaign
    /** Used for testing purposes. Sets the current date */
    now?: Date
    history: H.History
    displayNamespace: boolean
}

/**
 * An item in the list of campaigns.
 */
export const CampaignNode: React.FunctionComponent<CampaignNodeProps> = ({
    node,
    history,
    now = new Date(),
    displayNamespace,
}) => (
    <>
        <span className="campaign-node__separator" />
        {!node.closedAt && (
            <h2 className="m-0 campaign-node__badge">
                <span className="badge badge-success text-uppercase w-100">Open</span>
            </h2>
        )}
        {node.closedAt && (
            <h2 className="m-0 campaign-node__badge">
                <span className="badge badge-danger text-uppercase w-100">Closed</span>
            </h2>
        )}
        <div className="campaign-node__content">
            <div className="m-0 d-flex align-items-baseline">
                <h3 className="m-0 d-inline-block">
                    {displayNamespace && (
                        <>
                            <Link
                                className="text-muted test-campaign-namespace-link"
                                to={`${node.namespace.url}/campaigns`}
                            >
                                {node.namespace.namespaceName}
                            </Link>
                            <span className="text-muted d-inline-block mx-1">/</span>
                        </>
                    )}
                    <Link className="test-campaign-link" to={node.url}>
                        {node.name}
                    </Link>
                </h3>
                <small className="ml-2 text-muted">
                    created{' '}
                    <span data-tooltip={<Timestamp date={node.createdAt} />}>
                        {formatDistance(parseISO(node.createdAt), now)} ago
                    </span>
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
        <ChangesetStatusOpen label={<span className="text-muted">{node.changesets.stats.open} open</span>} />
        <ChangesetStatusClosed label={<span className="text-muted">{node.changesets.stats.closed} closed</span>} />
        <ChangesetStatusMerged label={<span className="text-muted">{node.changesets.stats.merged} merged</span>} />
    </>
)
