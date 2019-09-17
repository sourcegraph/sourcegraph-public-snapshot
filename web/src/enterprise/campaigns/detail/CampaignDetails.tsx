import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { useCampaignByID } from './useCampaignByID'
import { UserAvatar } from '../../../user/UserAvatar'
import { Timestamp } from '../../../components/time/Timestamp'
import { CampaignsIcon } from '../icons'
import { ChangesetList } from './changesets/ChangesetList'
import {
    changesetStatusColorClasses,
    changesetReviewStateColors,
    changesetStageLabels,
} from './changesets/presentation'
import { Link } from '../../../../../shared/src/components/Link'
import { groupBy } from 'lodash'

interface Props {
    /** The campaign ID. */
    campaignID: GQL.ID
}

const changesetStages: (GQL.ChangesetState | GQL.ChangesetReviewState)[] = [
    GQL.ChangesetState.MERGED,
    GQL.ChangesetState.CLOSED,
    GQL.ChangesetReviewState.APPROVED,
    GQL.ChangesetReviewState.CHANGES_REQUESTED,
    GQL.ChangesetReviewState.PENDING,
]
const changesetStageColors: Record<GQL.ChangesetReviewState | GQL.ChangesetState, string> = {
    ...changesetReviewStateColors,
    ...changesetStatusColorClasses,
}

/**
 * The area for a single campaign.
 */
export const CampaignDetails: React.FunctionComponent<Props> = ({ campaignID }) => {
    const campaign = useCampaignByID(campaignID)

    if (campaign === undefined) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (campaign === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }

    const changeSetCount = campaign.changesets.nodes.length

    const changesetsByStage = groupBy(campaign.changesets.nodes, changeset =>
        // For open changesets, group by review state
        changeset.state !== GQL.ChangesetState.OPEN ? changeset.state : changeset.reviewState
    )

    return (
        <>
            <PageTitle title={campaign.name} />
            <h2>
                <CampaignsIcon className="icon-inline" /> {campaign.namespace.namespaceName}
                <span className="text-muted d-inline-block mx-2">/</span>
                {campaign.name}
            </h2>
            <div className="card mb-3">
                <div className="card-header">
                    <strong>
                        <UserAvatar user={campaign.author} className="icon-inline" /> {campaign.author.username}
                    </strong>{' '}
                    started <Timestamp date={campaign.createdAt} />
                </div>
                <div className="card-body">{campaign.description}</div>
            </div>
            <h3>
                Changesets <span className="badge badge-secondary badge-pill">{campaign.changesets.nodes.length}</span>
            </h3>
            {changeSetCount > 0 && (
                <div>
                    <div className="progress rounded mb-2">
                        {changesetStages.map(stage => {
                            const changesetsInStage = changesetsByStage[stage] || []
                            const count = changesetsInStage.length
                            return (
                                count > 0 && (
                                    <div
                                        // Needed for dynamic width
                                        // eslint-disable-next-line react/forbid-dom-props
                                        style={{ width: (count / changeSetCount) * 100 + '%' }}
                                        className={`progress-bar bg-${changesetStageColors[stage]}`}
                                        role="progressbar"
                                        aria-valuemin={0}
                                        aria-valuenow={count}
                                        aria-valuemax={changeSetCount}
                                        key={stage}
                                    >
                                        {count} {changesetStageLabels[stage]}
                                    </div>
                                )
                            )
                        })}
                    </div>
                </div>
            )}
            <ChangesetList changesets={campaign.changesets.nodes} />
            <p className="mt-2">
                Use the <Link to="/api/console">GraphQL API</Link> to add changesets to this campaign (
                <code>createChangeset</code> and <code>addChangesetToCampaign</code>)
            </p>
        </>
    )
}
