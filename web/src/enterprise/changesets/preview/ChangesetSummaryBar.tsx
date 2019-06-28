import React from 'react'
import { Link } from 'react-router-dom'
import { RepositoryIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ActionsIcon, DiffIcon, GitCommitIcon } from '../../../util/octicons'
import { ChecklistIcon } from '../../checklists/icons'
import { ThreadSettings } from '../../threads/settings'

interface Props {
    thread: GQL.IDiscussionThread
    xchangeset: GQL.IChangeset
    threadSettings: ThreadSettings

    className?: string
}

interface SummaryItem {
    noun: string
    pluralNoun?: string
    icon: React.ComponentType<{ className?: string }>
    count: number | ((changeset: GQL.IChangeset, threadSettings: ThreadSettings) => number) | null
}

export const countChangesetActions = (c: GQL.IChangeset, s: ThreadSettings) =>
    s.changesetActionDescriptions ? s.changesetActionDescriptions.length : 0

export const countChangesetFilesChanged = (c: GQL.IChangeset) =>
    c.repositoryComparisons.reduce((n, c) => n + (c.fileDiffs.totalCount || 0), 0)

export const countChangesetCommits = (c: GQL.IChangeset) => c.commits.length

const ITEMS: SummaryItem[] = [
    {
        noun: 'action',
        icon: ActionsIcon,
        count: countChangesetActions,
    },
    { noun: 'Review task', icon: ChecklistIcon, count: null },
    {
        noun: 'repository affected',
        pluralNoun: 'repositories affected',
        icon: RepositoryIcon,
        count: c => c.repositories.length,
    },
    { noun: 'commit', icon: GitCommitIcon, count: countChangesetCommits },
    {
        noun: 'file changed',
        pluralNoun: 'files changed',
        icon: DiffIcon,
        count: countChangesetFilesChanged,
    },
]

/**
 * A bar that summarizes the contents and impact of a changeset.
 */
export const ChangesetSummaryBar: React.FunctionComponent<Props> = ({
    xchangeset,
    threadSettings,
    className = '',
    ...props
}) => (
    <nav className={`changeset-summary-bar border ${className}`}>
        <ul className="nav w-100">
            {ITEMS.map(({ icon: Icon, ...item }, i) => {
                const count =
                    typeof item.count === 'number'
                        ? item.count
                        : typeof item.count === 'function'
                        ? item.count(xchangeset, threadSettings)
                        : null
                return (
                    <li key={i} className="nav-item flex-1 text-center">
                        <Link to="TODO!(sqs)" className="nav-link">
                            <Icon className="icon-inline text-muted" /> <strong>{count}</strong>{' '}
                            <span className="text-muted">{pluralize(item.noun, count || 0, item.pluralNoun)}</span>
                        </Link>
                    </li>
                )
            })}
        </ul>
    </nav>
)
