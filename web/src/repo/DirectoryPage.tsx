import BranchIcon from '@sourcegraph/icons/lib/Branch'
import CommitIcon from '@sourcegraph/icons/lib/Commit'
import { Folder as FolderIcon } from '@sourcegraph/icons/lib/Folder'
import HistoryIcon from '@sourcegraph/icons/lib/History'
import { Loader } from '@sourcegraph/icons/lib/Loader'
import { Repo as RepositoryIcon } from '@sourcegraph/icons/lib/Repo'
import TagIcon from '@sourcegraph/icons/lib/Tag'
import UserIcon from '@sourcegraph/icons/lib/User'
import escapeRegexp from 'escape-string-regexp'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { makeRepoURI } from '.'
import { gql, queryGraphQL } from '../backend/graphql'
import { Form } from '../components/Form'
import { PageTitle } from '../components/PageTitle'
import { displayRepoPath } from '../components/RepoFileLink'
import { searchQueryForRepoRev } from '../search'
import { submitSearch } from '../search/helpers'
import { QueryInput } from '../search/QueryInput'
import { SearchButton } from '../search/SearchButton'
import { SearchHelp } from '../search/SearchHelp'
import { eventLogger } from '../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { memoizeObservable } from '../util/memoize'
import { toPrettyBlobURL, toRepoURL, toTreeURL } from '../util/url'
import { GitCommitNode } from './commits/GitCommitNode'
import { FilteredGitCommitConnection, gitCommitFragment } from './commits/RepositoryCommitsPage'

const DirectoryEntry: React.SFC<{
    isDir: boolean
    name: string
    parentPath: string
    repoPath: string
    rev?: string
}> = ({ isDir, name, parentPath, repoPath, rev }) => {
    const filePath = parentPath ? parentPath + '/' + name : name
    return (
        <Link
            to={(isDir ? toTreeURL : toPrettyBlobURL)({
                repoPath,
                rev,
                filePath,
            })}
            className="directory-entry"
            title={filePath}
        >
            {name}
            {isDir && '/'}
        </Link>
    )
}

export const fetchTree = memoizeObservable(
    (ctx: { repoPath: string; commitID: string; filePath: string }): Observable<GQL.ITree> =>
        queryGraphQL(
            gql`
                query Tree($repoPath: String!, $commitID: String!, $filePath: String!) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            tree(path: $filePath) {
                                directories {
                                    name
                                }
                                files {
                                    name
                                }
                            }
                        }
                    }
                }
            `,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (!data || errors || !data.repository || !data.repository.commit || !data.repository.commit.tree) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.tree
            })
        ),
    makeRepoURI
)

const fetchTreeCommits = memoizeObservable(
    (args: { repo: GQLID; revspec: string; first?: number; filePath?: string }): Observable<GQL.IGitCommitConnection> =>
        queryGraphQL(
            gql`
                query TreeCommits($repo: ID!, $revspec: String!, $first: Int, $filePath: String) {
                    node(id: $repo) {
                        ... on Repository {
                            commit(rev: $revspec) {
                                ancestors(first: $first, path: $filePath) {
                                    nodes {
                                        ...GitCommitFields
                                    }
                                    pageInfo {
                                        hasNextPage
                                    }
                                }
                            }
                        }
                    }
                }
                ${gitCommitFragment}
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const repo = data.node as GQL.IRepository
                if (!repo.commit || !repo.commit.ancestors || !repo.commit.ancestors.nodes) {
                    throw createAggregateError(errors)
                }
                return repo.commit.ancestors
            })
        ),
    args => `${args.repo}:${args.revspec}:${args.first}:${args.filePath}`
)

interface Props {
    repoPath: string
    repoID: GQLID
    repoDescription: string
    // filePath is a directory path in DirectoryPage. We call it filePath for consistency elsewhere.
    filePath: string
    commitID: string
    rev?: string
    isLightTheme: boolean

    location: H.Location
    history: H.History
}

interface State {
    /** This directory's tree, or an error. Undefined while loading. */
    treeOrError?: GQL.ITree | ErrorLike

    /**
     * The value of the search query input field.
     */
    query: string
}

/** Feature flag for showing the contributors link. */
const showContributors = localStorage.getItem('contributors') !== null

export class DirectoryPage extends React.PureComponent<Props, State> {
    public state: State = { query: '' }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    private logViewEvent(props: Props): void {
        if (props.filePath === '') {
            eventLogger.logViewEvent('Repository')
        } else {
            eventLogger.logViewEvent('Directory')
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(
                        (x, y) =>
                            x.repoPath === y.repoPath &&
                            x.rev === y.rev &&
                            x.commitID === y.commitID &&
                            x.filePath === y.filePath
                    ),
                    tap(props => this.logViewEvent(props)),
                    switchMap(props =>
                        fetchTree(props).pipe(
                            catchError(err => [asError(err)]),
                            map(c => ({ treeOrError: c })),
                            startWith<Pick<State, 'treeOrError'>>({ treeOrError: undefined })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private getQueryPrefix(): string {
        let queryPrefix = searchQueryForRepoRev(this.props.repoPath, this.props.rev)
        if (this.props.filePath) {
            queryPrefix += `file:^${escapeRegexp(this.props.filePath)}/ `
        }
        return queryPrefix
    }

    public render(): JSX.Element | null {
        return (
            <div className="directory-page">
                <PageTitle key="page-title" title={this.getPageTitle()} />
                {this.props.filePath ? (
                    <header>
                        <h2 className="directory-page__title">
                            <FolderIcon className="icon-inline" /> {this.props.filePath}
                        </h2>
                    </header>
                ) : (
                    <header>
                        <h2 className="directory-page__title">
                            <RepositoryIcon className="icon-inline" /> {displayRepoPath(this.props.repoPath)}
                        </h2>
                        {this.props.repoDescription && <p>{this.props.repoDescription}</p>}
                        <div className="btn-group mb-3">
                            <Link
                                className="btn btn-secondary"
                                to={`${toRepoURL({ repoPath: this.props.repoPath, rev: this.props.rev })}/-/commits`}
                            >
                                <CommitIcon className="icon-inline" /> Commits
                            </Link>
                            <Link className="btn btn-secondary" to={`/${this.props.repoPath}/-/branches`}>
                                <BranchIcon className="icon-inline" /> Branches
                            </Link>
                            <Link className="btn btn-secondary" to={`/${this.props.repoPath}/-/tags`}>
                                <TagIcon className="icon-inline" /> Tags
                            </Link>
                            <Link
                                className="btn btn-secondary"
                                to={
                                    this.props.rev
                                        ? `/${this.props.repoPath}/-/compare/...${this.props.rev}`
                                        : `/${this.props.repoPath}/-/compare`
                                }
                            >
                                <HistoryIcon className="icon-inline" /> Compare
                            </Link>
                            {showContributors && (
                                <Link
                                    className={`btn btn-outline-${this.props.isLightTheme ? 'dark' : 'light'}`}
                                    to={`/${this.props.repoPath}/-/stats/contributors`}
                                >
                                    <UserIcon className="icon-inline" /> Contributors
                                </Link>
                            )}
                        </div>
                    </header>
                )}

                <section className="directory-page__section">
                    <h3 className="directory-page__section-header">
                        Search in this {this.props.filePath ? 'directory' : 'repository'}
                    </h3>
                    <Form className="directory-page__section-search" onSubmit={this.onSubmit}>
                        <QueryInput
                            value={this.state.query}
                            onChange={this.onQueryChange}
                            prependQueryForSuggestions={this.getQueryPrefix()}
                            autoFocus={true}
                            location={this.props.location}
                            history={this.props.history}
                            placeholder=""
                        />
                        <SearchButton />
                        <SearchHelp />
                    </Form>
                </section>
                {this.state.treeOrError === undefined && (
                    <div>
                        <Loader className="icon-inline directory-page__entries-loader" /> Loading files and directories
                    </div>
                )}
                {this.state.treeOrError !== undefined &&
                    (isErrorLike(this.state.treeOrError) ? (
                        <div className="alert alert-danger">
                            <p>Unable to list directory contents</p>
                            {this.state.treeOrError.message && (
                                <div>
                                    <pre>{this.state.treeOrError.message.slice(0, 100)}</pre>
                                </div>
                            )}
                        </div>
                    ) : (
                        <>
                            {this.state.treeOrError.directories.length > 0 && (
                                <section className="directory-page__section">
                                    <h3 className="directory-page__section-header">Directories</h3>
                                    <div className="directory-page__entries directory-page__entries-directories">
                                        {this.state.treeOrError.directories.map((e, i) => (
                                            <DirectoryEntry
                                                key={i}
                                                isDir={true}
                                                name={e.name}
                                                parentPath={this.props.filePath}
                                                repoPath={this.props.repoPath}
                                                rev={this.props.rev}
                                            />
                                        ))}
                                    </div>
                                </section>
                            )}
                            {this.state.treeOrError.files.length > 0 && (
                                <section className="directory-page__section">
                                    <h3 className="directory-page__section-header">Files</h3>
                                    <div className="directory-page__entries directory-page__entries-files">
                                        {this.state.treeOrError.files.map((e, i) => (
                                            <DirectoryEntry
                                                key={i}
                                                isDir={false}
                                                name={e.name}
                                                parentPath={this.props.filePath}
                                                repoPath={this.props.repoPath}
                                                rev={this.props.rev}
                                            />
                                        ))}
                                    </div>
                                </section>
                            )}
                        </>
                    ))}
                <div className="directory-page__section">
                    <h3 className="directory-page__section-header">Changes</h3>
                    <FilteredGitCommitConnection
                        className="mt-2 directory-page__section--commits"
                        listClassName="list-group list-group-flush"
                        noun="commit in this tree"
                        pluralNoun="commits in this tree"
                        queryConnection={this.queryCommits}
                        nodeComponent={GitCommitNode}
                        nodeComponentProps={{
                            repoName: this.props.repoPath,
                            className: 'list-group-item',
                            compact: true,
                        }}
                        updateOnChange={`${this.props.repoPath}:${this.props.rev}:${this.props.filePath}`}
                        defaultFirst={7}
                        history={this.props.history}
                        shouldUpdateURLQuery={false}
                        hideFilter={true}
                        location={this.props.location}
                    />
                </div>
            </div>
        )
    }

    private onQueryChange = (query: string) => this.setState({ query })

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(
            this.props.history,
            { query: this.getQueryPrefix() + this.state.query },
            this.props.filePath ? 'tree' : 'repo'
        )
    }

    private getPageTitle(): string {
        const repoPathSplit = this.props.repoPath.split('/')
        const repoStr = repoPathSplit.length > 2 ? repoPathSplit.slice(1).join('/') : this.props.repoPath
        if (this.props.filePath) {
            const fileOrDir = this.props.filePath.split('/').pop()
            return `${fileOrDir} - ${repoStr}`
        }
        return `${repoStr}`
    }

    private queryCommits = (args: { first?: number }) =>
        fetchTreeCommits({
            ...args,
            repo: this.props.repoID,
            revspec: this.props.commitID,
            filePath: this.props.filePath,
        })
}
