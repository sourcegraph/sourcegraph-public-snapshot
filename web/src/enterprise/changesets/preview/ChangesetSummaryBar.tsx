import React from 'react'
import { Link } from 'react-router-dom'
import { RepositoryIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../../shared/src/util/strings'
import { TasksIcon } from '../../tasks/icons'
import { ThreadSettings } from '../../threads/settings'
import { ActionsIcon, DiffIcon, GitCommitIcon } from '../icons'

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
    count: number | ((changeset: GQL.IChangeset, threadSettings: ThreadSettings) => number)
}

const ITEMS: SummaryItem[] = [
    {
        noun: 'action',
        icon: ActionsIcon,
        count: (_c, s) => (s.changesetActionDescriptions ? s.changesetActionDescriptions.length : 0),
    },
    { noun: 'review task', icon: TasksIcon, count: 7 },
    {
        noun: 'repository affected',
        pluralNoun: 'repositories affected',
        icon: RepositoryIcon,
        count: c => c.repositories.length,
    },
    { noun: 'commit', icon: GitCommitIcon, count: c => c.commits.length },
    {
        noun: 'file changed',
        pluralNoun: 'files changed',
        icon: DiffIcon,
        count: c => c.repositoryComparisons.reduce((n, c) => n + (c.fileDiffs.totalCount || 0), 0),
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
                const count = typeof item.count === 'number' ? item.count : item.count(xchangeset, threadSettings)
                return (
                    <li key={i} className="nav-item flex-1 text-center">
                        <Link to="TODO!(sqs)" className="nav-link">
                            <Icon className="icon-inline text-muted" /> <strong>{count}</strong>{' '}
                            <span className="text-muted">{pluralize(item.noun, count, item.pluralNoun)}</span>
                        </Link>
                    </li>
                )
            })}
        </ul>
    </nav>
)
