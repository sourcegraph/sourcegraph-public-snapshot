import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import React from 'react'
import { RepositoryIcon } from '../../../../../../shared/src/components/icons'
import { RepoLink } from '../../../../../../shared/src/components/RepoLink'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { GitCommitNode } from '../../../../repo/commits/GitCommitNode'
import { DiffStat } from '../../../../repo/compare/DiffStat'
import { GitCommitIcon } from '../../../../util/octicons'
import { useCampaignRepositories } from './useCampaignRepositories'
import { Link } from 'react-router-dom'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'

interface Props {
    campaign: Pick<GQL.ICampaign, 'id'>

    showCommits?: boolean

    className?: string
}

const LOADING = 'loading' as const

/**
 * A list of repositories affected by a campaign.
 */
export const CampaignRepositoriesList: React.FunctionComponent<Props> = ({ campaign, showCommits, className = '' }) => {
    const repositories = useCampaignRepositories(campaign)
    return (
        <div className={`campaign-repositories-list ${className}`}>
            <ul className="list-group mb-4">
                {repositories === LOADING ? (
                    <LoadingSpinner className="icon-inline mt-3" />
                ) : isErrorLike(repositories) ? (
                    <div className="alert alert-danger mt-3">{repositories.message}</div>
                ) : (
                    repositories.map((c, i) => (
                        <li key={i} className="list-group-item">
                            <div className="d-flex align-items-center">
                                <Link to={c.baseRepository.url} className="mr-3">
                                    <RepositoryIcon className="icon-inline" /> {displayRepoName(c.baseRepository.name)}
                                </Link>
                                <span className="text-muted d-inline-flex align-items-center">
                                    {c.range.baseRevSpec.expr.replace(/^refs\/heads\//, '')}{' '}
                                    <DotsHorizontalIcon className="icon-inline small" />{' '}
                                    {c.range.headRevSpec.expr.replace(/^refs\/heads\//, '')}
                                </span>
                                <div className="flex-1"></div>
                                {!showCommits && (
                                    <small className="mr-3">
                                        <GitCommitIcon className="icon-inline" /> {c.commits.nodes.length}{' '}
                                        {pluralize('commit', c.commits.nodes.length)}
                                    </small>
                                )}
                                <DiffStat {...c.fileDiffs.diffStat} />
                            </div>
                            {showCommits && (
                                <ul className="list-group">
                                    {c.commits.nodes.map((commit, i) => (
                                        <li
                                            key={i}
                                            className="list-group-item border-0 d-flex align-items-start pb-0 px-0 border-left ml-4 pl-4"
                                        >
                                            <GitCommitIcon className="icon-inline mr-3 text-muted" />
                                            <GitCommitNode node={commit} compact={true} className="p-0 flex-1" />
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </li>
                    ))
                )}
            </ul>
        </div>
    )
}
