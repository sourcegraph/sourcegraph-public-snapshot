import H from 'history'
import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import React from 'react'
import { RepositoryIcon } from '../../../../../../shared/src/components/icons'
import { RepoLink } from '../../../../../../shared/src/components/RepoLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { GitCommitNode } from '../../../../repo/commits/GitCommitNode'
import { DiffStat } from '../../../../repo/compare/DiffStat'
import { GitRefTag } from '../../../../repo/GitRefTag'
import { GitCommitIcon } from '../../../../util/octicons'
import { ThreadSettings } from '../../../threads/settings'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    thread: GQL.IDiscussionThread
    xchangeset: GQL.IChangeset
    threadSettings: ThreadSettings

    showCommits?: boolean

    className?: string
    location: H.Location
    history: H.History
    isLightTheme: boolean
}

/**
 * A list of affected repositories in a changeset.
 */
export const ChangesetRepositoriesList: React.FunctionComponent<Props> = ({
    xchangeset,
    showCommits,
    className = '',
}) => (
    <div className={`changeset-repositories-list ${className}`}>
        <ul className="list-group mb-4">
            {xchangeset.repositoryComparisons.map((c, i) => (
                <li key={i} className="list-group-item">
                    <div className="d-flex align-items-center">
                        <RepoLink
                            key={c.baseRepository.id}
                            repoName={c.baseRepository.name}
                            to={c.baseRepository.url}
                            icon={RepositoryIcon}
                            className="mr-3"
                        />
                        <span className="text-muted d-inline-flex align-items-center">
                            {c.range.baseRevSpec.expr} <DotsHorizontalIcon className="icon-inline small" />{' '}
                            {c.range.headRevSpec.expr}
                        </span>
                        <div className="flex-1"></div>
                        <small className="mr-3">
                            <GitCommitIcon className="icon-inline" /> {c.commits.nodes.length}{' '}
                            {pluralize('commit', c.commits.nodes.length)}
                        </small>
                        <DiffStat {...c.fileDiffs.diffStat} />
                    </div>
                    {showCommits && (
                        <ul className="list-group ml-6">
                            {c.commits.nodes.map((commit, i) => (
                                <li key={i} className="list-group-item border-0 d-flex align-items-start">
                                    <GitCommitIcon className="icon-inline mr-3 text-muted" />
                                    <GitCommitNode
                                        repoName={c.baseRepository.name}
                                        node={commit}
                                        compact={true}
                                        className="p-0 flex-1"
                                    />
                                </li>
                            ))}
                        </ul>
                    )}
                </li>
            ))}
        </ul>
    </div>
)
