import CircleChevronLeft from '@sourcegraph/icons/lib/CircleChevronLeft'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { replaceRevisionInURL } from '.'
import { gql, queryGraphQL } from '../backend/graphql'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { eventLogger } from '../tracking/eventLogger'
import { memoizeObservable } from '../util/memoize'

const fetchGitRefs = memoizeObservable(
    (args: {
        repo: GQLID
        first?: number
        query?: string
        type?: GQL.IGitRefTypeEnum
    }): Observable<GQL.IGitRefConnection> =>
        queryGraphQL(
            gql`
                query RepositoryGitRefs($repo: ID!, $first: Int, $query: String, $type: GitRefType) {
                    node(id: $repo) {
                        ... on Repository {
                            gitRefs(first: $first, query: $query, type: $type) {
                                nodes {
                                    id
                                    name
                                    displayName
                                    abbrevName
                                    type
                                }
                                totalCount
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || !(data.node as GQL.IRepository).gitRefs) {
                    throw Object.assign(
                        'Could not fetch repository Git refs: ' +
                            new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                return (data.node as GQL.IRepository).gitRefs
            })
        ),
    x => JSON.stringify(x)
)

const fetchRepositoryCommits = memoizeObservable(
    (args: { repo: GQLID; rev?: string; first?: number; query?: string }): Observable<GQL.IGitCommitConnection> =>
        queryGraphQL(
            gql`
                query RepositoryGitCommit($repo: ID!, $first: Int, $rev: String!, $query: String) {
                    node(id: $repo) {
                        ... on Repository {
                            commit(rev: $rev) {
                                ancestors(first: $first, query: $query) {
                                    nodes {
                                        id
                                        oid
                                        abbreviatedOID
                                        author {
                                            person {
                                                name
                                                avatarURL
                                            }
                                            date
                                        }
                                        subject
                                    }
                                    pageInfo {
                                        hasNextPage
                                    }
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.node ||
                    !(data.node as GQL.IRepository).commit ||
                    !(data.node as GQL.IRepository).commit!.ancestors
                ) {
                    throw Object.assign(
                        'Could not fetch commits: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                return (data.node as GQL.IRepository).commit!.ancestors
            })
        ),
    x => JSON.stringify(x)
)

interface GitRefNodeProps {
    node: GQL.IGitRef

    defaultBranch: string | undefined
    currentRev: string | undefined

    location: H.Location
}

export const GitRefNode: React.SFC<GitRefNodeProps> = ({ node, defaultBranch, currentRev, location }) => {
    let isCurrent: boolean
    if (currentRev) {
        isCurrent = node.name === currentRev || node.abbrevName === currentRev
    } else {
        isCurrent = node.name === `refs/heads/${defaultBranch}`
    }

    return (
        <li key={node.id} className="popover__node">
            <Link
                to={replaceRevisionInURL(location.pathname + location.search + location.hash, node.abbrevName)}
                className={`popover__node-link ${isCurrent ? 'popover__node-link--active' : ''}`}
            >
                {node.displayName}
                {isCurrent && (
                    <CircleChevronLeft className="icon-inline popover__node-link-icon" data-tooltip="Current" />
                )}
            </Link>
        </li>
    )
}

interface GitCommitNodeProps {
    node: GQL.IGitCommit

    currentCommitID: string | undefined

    location: H.Location
}

export const GitCommitNode: React.SFC<GitCommitNodeProps> = ({ node, currentCommitID, location }) => {
    const isCurrent = currentCommitID === (node.oid as string)
    return (
        <li key={node.oid} className="popover__node revisions-popover-git-commit-node">
            <Link
                to={replaceRevisionInURL(location.pathname + location.search + location.hash, node.oid as string)}
                className={`popover__node-link ${
                    isCurrent ? 'popover__node-link--active' : ''
                } revisions-popover-git-commit-node__link`}
            >
                <code className="revisions-popover-git-commit-node__oid" title={node.oid}>
                    {node.abbreviatedOID}
                </code>
                <span className="revisions-popover-git-commit-node__message">{(node.subject || '').slice(0, 200)}</span>
                {isCurrent && (
                    <CircleChevronLeft
                        className="icon-inline popover__node-link-icon revisions-popover-git-commit-node__icon"
                        data-tooltip="Current commit"
                    />
                )}
            </Link>
        </li>
    )
}

interface Props {
    repo: GQLID
    repoPath: string
    defaultBranch: string | undefined

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    currentCommitID?: string

    history: H.History
    location: H.Location
}

interface State {
    activeTab: RevisionsPopoverTabValue
}

type RevisionsPopoverTabValue = 'branches' | 'tags' | 'commits'

interface RevisionsPopoverTab {
    label: string
    value: RevisionsPopoverTabValue
    noun: string
    pluralNoun: string
    type?: GQL.IGitRefTypeEnum
}

class FilteredGitRefConnection extends FilteredConnection<
    GQL.IGitRef,
    Pick<GitRefNodeProps, 'defaultBranch' | 'currentRev' | 'location'>
> {}

class FilteredGitCommitConnection extends FilteredConnection<
    GQL.IGitCommit,
    Pick<GitCommitNodeProps, 'currentCommitID' | 'location'>
> {}

/**
 * A popover that displays a searchable list of revisions (grouped by type) for
 * the current repository.
 */
export class RevisionsPopover extends React.PureComponent<Props> {
    private static LAST_TAB_STORAGE_KEY = 'RevisionsPopover.lastTab'

    private static TABS: RevisionsPopoverTab[] = [
        { label: 'Branches', value: 'branches', noun: 'branch', pluralNoun: 'branches', type: 'GIT_BRANCH' },
        { label: 'Tags', value: 'tags', noun: 'tag', pluralNoun: 'tags', type: 'GIT_TAG' },
        { label: 'Commits', value: 'commits', noun: 'commit', pluralNoun: 'commits' },
    ]

    public state: State = {
        activeTab:
            (localStorage.getItem(RevisionsPopover.LAST_TAB_STORAGE_KEY) as RevisionsPopoverTabValue | null) ||
            'branches',
    }

    private static saveToLocalStorage(lastTab: RevisionsPopoverTabValue): void {
        localStorage.setItem(RevisionsPopover.LAST_TAB_STORAGE_KEY, lastTab)
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RevisionsPopover')
    }

    public render(): JSX.Element | null {
        const activeTab = RevisionsPopover.TABS.find(({ value }) => this.state.activeTab === value)!

        return (
            <div className="revisions-popover popover">
                <div className="revisions-popover__tabs">
                    {RevisionsPopover.TABS.map(({ label, value }) => (
                        <button
                            key={value}
                            className={`btn btn-link btn-sm revisions-popover__tab revisions-popover__tab--${
                                this.state.activeTab === value ? 'active' : 'inactive'
                            }`}
                            // tslint:disable-next-line:jsx-no-lambda
                            onClick={() =>
                                this.setState({ activeTab: value }, () => RevisionsPopover.saveToLocalStorage(value))
                            }
                        >
                            {label}
                        </button>
                    ))}
                </div>
                {activeTab.type ? (
                    <FilteredGitRefConnection
                        key={activeTab.value}
                        className="popover__content"
                        compact={true}
                        noun={activeTab.noun}
                        pluralNoun={activeTab.pluralNoun}
                        // tslint:disable-next-line:jsx-no-lambda
                        queryConnection={(args: FilteredConnectionQueryArgs) =>
                            fetchGitRefs({ ...args, repo: this.props.repo, type: activeTab.type })
                        }
                        nodeComponent={GitRefNode}
                        nodeComponentProps={
                            {
                                defaultBranch: this.props.defaultBranch,
                                currentRev: this.props.currentRev,
                                location: this.props.location,
                            } as Pick<GitRefNodeProps, 'defaultBranch' | 'currentRev' | 'location'>
                        }
                        defaultFirst={50}
                        autoFocus={true}
                        history={this.props.history}
                        location={this.props.location}
                        noUpdateURLQuery={true}
                        noSummaryIfAllNodesVisible={true}
                    />
                ) : (
                    <FilteredGitCommitConnection
                        key={activeTab.value}
                        className="popover__content"
                        compact={true}
                        noun={activeTab.noun}
                        pluralNoun={activeTab.pluralNoun}
                        // tslint:disable-next-line:jsx-no-lambda
                        queryConnection={(args: FilteredConnectionQueryArgs) =>
                            fetchRepositoryCommits({
                                ...args,
                                repo: this.props.repo,
                                rev: this.props.currentRev || this.props.defaultBranch,
                            })
                        }
                        nodeComponent={GitCommitNode}
                        nodeComponentProps={
                            {
                                currentCommitID: this.props.currentCommitID,
                                location: this.props.location,
                            } as Pick<GitCommitNodeProps, 'currentCommitID' | 'location'>
                        }
                        defaultFirst={15}
                        autoFocus={true}
                        history={this.props.history}
                        location={this.props.location}
                        noUpdateURLQuery={true}
                        noSummaryIfAllNodesVisible={true}
                    />
                )}
            </div>
        )
    }
}
