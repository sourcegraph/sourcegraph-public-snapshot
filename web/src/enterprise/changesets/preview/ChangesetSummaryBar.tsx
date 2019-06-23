import React from 'react'
import { Link } from 'react-router-dom'
import { RepositoryIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { TasksIcon } from '../../tasks/icons'
import { ThreadSettings } from '../../threads/settings'
import { ActionsIcon, DiffIcon, GitCommitIcon } from '../icons'

interface Props {
    thread: GQL.IDiscussionThread
    threadSettings: ThreadSettings

    className?: string
}

interface SummaryItem {
    noun: string
    pluralNoun?: string
    icon: React.ComponentType<{ className?: string }>
    count: number // TODO!(sqs) | ((thread: Pick<GQL.IDiscussionThread, 'id'>) => Promise<number|ErrorLike>)
}

const ITEMS: SummaryItem[] = [
    { noun: 'action', icon: ActionsIcon, count: 3 },
    { noun: 'review task', icon: TasksIcon, count: 7 },
    { noun: 'repository affected', pluralNoun: 'repositories affected', icon: RepositoryIcon, count: 31 },
    { noun: 'commit', icon: GitCommitIcon, count: 31 },
    { noun: 'file changed', pluralNoun: 'files changed', icon: DiffIcon, count: 62 },
]

/**
 * A bar that summarizes the contents and impact of a changeset.
 */
export const ChangesetSummaryBar: React.FunctionComponent<Props> = ({ className = '', ...props }) => (
    <nav className={`changeset-summary-bar border ${className}`}>
        <ul className="nav w-100">
            {ITEMS.map(({ icon: Icon, ...item }, i) => (
                <li key={i} className="nav-item flex-1 text-center">
                    <Link to="TODO!(sqs)" className="nav-link">
                        <Icon className="icon-inline text-muted" /> <strong>{item.count}</strong>{' '}
                        <span className="text-muted">{pluralize(item.noun, item.count, item.pluralNoun)}</span>
                    </Link>
                </li>
            ))}
        </ul>
    </nav>
)
