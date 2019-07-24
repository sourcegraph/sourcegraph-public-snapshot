import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { Timestamp } from '../../../../components/time/Timestamp'
import { Timeline } from '../../../../components/timeline/Timeline'
import { PersonLink } from '../../../../user/PersonLink'
import { ActionsIcon, GitPullRequestIcon } from '../../../../util/octicons'
import { ChecksIcon } from '../../../checks/icons'
import { ThreadStatusBadge } from '../../../threads/components/threadStatus/ThreadStatusBadge'
import { ThreadHeaderEditableTitle } from '../../../threads/detail/header/ThreadHeaderEditableTitle'
import { ThreadBreadcrumbs } from '../../../threads/detail/overview/ThreadBreadcrumbs'
import { ThreadDescription } from '../../../threads/detail/overview/ThreadDescription'
import { ThreadStatusButton } from '../../../threads/form/ThreadStatusButton'
import { ThreadSettings } from '../../../threads/settings'
import { countCampaignFilesChanged } from '../../preview/CampaignSummaryBar'
import { CampaignReviewLink } from '../changes/CampaignReviewLink'
import { CampaignReviewsList } from '../changes/CampaignReviewsList'
import { CampaignAreaContext } from './CampaignArea'

interface Props extends CampaignAreaContext {
    className?: string
    history: H.History
    location: H.Location
}

/**
 * The overview for a single campaign.
 */
export const CampaignOverview: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => (
    // TODO!(sqs): uses style from other component
    <div className={`thread-overview ${className || ''}`}>
        <ThreadBreadcrumbs thread={campaign} project={null /* TODO!(sqs) */} areaURL={areaURL} className="py-3" />
        <hr className="my-0" />
        <div className="d-flex align-items-center py-3">
            <ThreadStatusBadge thread={campaign} className="mr-3" />
            <div>
                <small>
                    Opened <Timestamp date={campaign.createdAt} /> by{' '}
                    <strong>
                        <PersonLink user={campaign.author} />
                    </strong>
                </small>
            </div>
            <div className="flex-1" />
            <ThreadStatusButton
                {...props}
                thread={campaign}
                onThreadUpdate={onThreadUpdate}
                className="ml-2"
                buttonClassName="btn-link btn-sm"
            />
        </div>
        <hr className="my-0" />
        <h2>{campaign.title}</h2>
        <ThreadHeaderEditableTitle
            {...props}
            thread={campaign}
            onThreadUpdate={onThreadUpdate}
            className="thread-overview__thread-title py-3" // TODO!(sqs): uses style from other component
        />
        <ThreadDescription {...props} thread={campaign} onThreadUpdate={onThreadUpdate} className="mb-4" />
        <Timeline className="align-items-stretch mb-4">
            {threadSettings.plan && (
                <div className="d-flex align-items-start bg-body border p-4 mb-5 position-relative">
                    <ActionsIcon className="mb-0 mr-3" />
                    <Link to={`${areaURL}/operations`} className="stretched-link text-body">
                        {threadSettings.plan.operations.length}{' '}
                        {pluralize('operation', threadSettings.plan.operations.length)} applied{' '}
                        <span className="text-muted">
                            {' '}
                            &mdash; {countCampaignFilesChanged(xchangeset)}{' '}
                            {pluralize('file', countCampaignFilesChanged(xchangeset))} changed in{' '}
                            {xchangeset.repositories.length}{' '}
                            {pluralize('repository', xchangeset.repositories.length, 'repositories')}
                        </span>
                    </Link>
                </div>
            )}
            {threadSettings.relatedPRs && (
                <div className="d-flex align-items-center bg-body border p-4 mb-5">
                    <GitPullRequestIcon className="icon-inline mb-0 mr-3" />
                    Waiting for approval on {threadSettings.relatedPRs.length}{' '}
                    {pluralize('pull request', threadSettings.relatedPRs.length)}: <span className="mr-2"></span>
                    {threadSettings.relatedPRs.map((link, i) => (
                        <CampaignReviewLink
                            key={i}
                            link={link}
                            showRepositoryName={true}
                            showIcon={true}
                            className="p-2 mr-2"
                            iconClassName="small text-success"
                        />
                    ))}
                </div>
            )}
            {/* <div className="d-flex align-items-start bg-body border p-4 mb-5 position-relative">
                <ChecksIcon className="mb-0 mr-3" />
                <Link to={`${areaURL}/tasks`} className="stretched-link text-body">
                    Review is not complete
                </Link>
            </div> */}
            <div className="d-flex align-items-center bg-body border p-4">
                <button type="button" className="btn btn-secondary text-muted mr-4" disabled={true}>
                    <GitPullRequestIcon className="mb-0 mr-3" />
                    Merge all
                </button>
                <span className="text-muted">Requires all pull requests to be approved</span>
            </div>
        </Timeline>
    </div>
)
