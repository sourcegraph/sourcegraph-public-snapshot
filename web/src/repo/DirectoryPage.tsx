import { Folder as FolderIcon } from '@sourcegraph/icons/lib/Folder'
import { Loader } from '@sourcegraph/icons/lib/Loader'
import { Repo as RepositoryIcon } from '@sourcegraph/icons/lib/Repo'
import formatDistance from 'date-fns/formatDistance'
import escapeRegexp from 'escape-string-regexp'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { makeRepoURI } from '.'
import { gql, queryGraphQL } from '../backend/graphql'
import { displayRepoPath } from '../components/Breadcrumb'
import { PageTitle } from '../components/PageTitle'
import { submitSearch } from '../search/helpers'
import { QueryInput } from '../search/QueryInput'
import { SearchButton } from '../search/SearchButton'
import { SearchHelp } from '../search/SearchHelp'
import { UserAvatar } from '../user/UserAvatar'
import { memoizeObservable } from '../util/memoize'
import { parseCommitDateString } from '../util/time'
import { externalCommitURL, toPrettyBlobURL, toTreeURL } from '../util/url'
import { searchQueryForRepoRev } from './RepoContainer'

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

export const fetchTreeAndCommits = memoizeObservable(
    (ctx: {
        repoPath: string
        commitID: string
        filePath: string
    }): Observable<{ tree: GQL.ITree; commits: GQL.ICommitInfo[] }> =>
        queryGraphQL(
            gql`
                query fetchTreeAndCommits($repoPath: String, $commitID: String, $filePath: String) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            commit {
                                tree(path: $filePath) {
                                    directories {
                                        name
                                    }
                                    files {
                                        name
                                    }
                                }
                                file(path: $filePath) {
                                    commits {
                                        abbreviatedOID
                                        message
                                        author {
                                            person {
                                                name
                                                avatarURL
                                            }
                                            date
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            `,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    errors ||
                    !data.repository ||
                    !data.repository.commit.commit ||
                    !data.repository.commit.commit.tree ||
                    !data.repository.commit.commit.file ||
                    !data.repository.commit.commit.file.commits
                ) {
                    throw Object.assign(
                        'Could not fetch tree and commits: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                return { tree: data.repository.commit.commit.tree, commits: data.repository.commit.commit.file.commits }
            })
        ),
    makeRepoURI
)

interface Props {
    repoPath: string
    repoDescription: string
    // filePath is a directory path in DirectoryPage. We call it filePath for consistency elsewhere.
    filePath: string
    commitID: string
    rev?: string

    location: H.Location
    history: H.History
}

interface State {
    loading: boolean
    tree?: GQL.ITree
    commits?: GQL.ICommitInfo[]
    errorDescription?: string

    /**
     * The value of the search query input field.
     */
    query: string
}

export class DirectoryPage extends React.PureComponent<Props, State> {
    public state: State = {
        loading: false,
        query: '',
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    constructor(props: Props) {
        super(props)
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
                    tap(() =>
                        this.setState({
                            loading: true,
                            tree: undefined,
                            commits: undefined,
                            errorDescription: undefined,
                        })
                    ),
                    switchMap(props =>
                        fetchTreeAndCommits(props).pipe(
                            catchError(err => {
                                this.setState({ loading: false, errorDescription: err })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    ({ tree, commits }) => this.setState({ tree, commits, loading: false }),
                    err => console.error(err)
                )
        )
    }

    public componentDidMount(): void {
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
                        <h1 className="directory-page__title">
                            <FolderIcon className="icon-inline" /> {this.props.filePath}
                        </h1>
                    </header>
                ) : (
                    <header>
                        <h1 className="directory-page__title">
                            <RepositoryIcon className="icon-inline" /> {displayRepoPath(this.props.repoPath)}
                        </h1>
                        {this.props.repoDescription && <p>{this.props.repoDescription}</p>}
                    </header>
                )}

                <section className="directory-page__section">
                    <h3 className="directory-page__section-header">
                        Search in this {this.props.filePath ? 'directory' : 'repository'}
                    </h3>
                    <form className="directory-page__section-search" onSubmit={this.onSubmit}>
                        <QueryInput
                            value={this.state.query}
                            onChange={this.onQueryChange}
                            prependQueryForSuggestions={this.getQueryPrefix()}
                            autoFocus={'cursor-at-end'}
                            location={this.props.location}
                            history={this.props.history}
                            placeholder=""
                        />
                        <SearchButton />
                        <SearchHelp />
                    </form>
                </section>
                {this.state.loading && <Loader className="icon-inline directory-page__entries-loader" />}
                {this.state.errorDescription && (
                    <div className="alert alert-danger">
                        <p>Error fetching directory information</p>
                        <div>
                            <pre>{this.state.errorDescription.slice(0, 100)}</pre>
                        </div>
                    </div>
                )}
                {this.state.tree &&
                    this.state.tree.directories.length > 0 && (
                        <section className="directory-page__section">
                            <h3 className="directory-page__section-header">Directories</h3>
                            <div className="directory-page__entries directory-page__entries-directories">
                                {this.state.tree.directories.map((e, i) => (
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
                {this.state.tree &&
                    this.state.tree.files.length > 0 && (
                        <section className="directory-page__section">
                            <h3 className="directory-page__section-header">Files</h3>
                            <div className="directory-page__entries directory-page__entries-files">
                                {this.state.tree.files.map((e, i) => (
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
                {this.state.commits &&
                    this.state.commits.length > 0 && (
                        <section className="directory-page__section">
                            <h3 className="directory-page__section-header">Recent changes</h3>
                            {this.props.rev && (
                                <div>
                                    From <code>{this.props.rev}</code>
                                </div>
                            )}
                            <table className="directory-page__section-commits table">
                                <tbody>
                                    {this.state.commits.map((c, i) => (
                                        <tr key={i} className="directory-page__commit">
                                            <td className="directory-page__commit-id" title={c.abbreviatedOID}>
                                                <a href={externalCommitURL(this.props.repoPath, this.props.commitID)}>
                                                    <code>{c.abbreviatedOID}</code>
                                                </a>
                                            </td>
                                            <td className="directory-page__commit-author">
                                                {c.author.person && <UserAvatar user={c.author.person} />}{' '}
                                                {c.author.person && c.author.person.name}
                                            </td>
                                            <td className="directory-page__commit-date" title={c.author.date}>
                                                {formatDistance(parseCommitDateString(c.author.date), new Date(), {
                                                    addSuffix: true,
                                                })}
                                            </td>
                                            <td className="directory-page__commit-message" title={c.message}>
                                                {c.message}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </section>
                    )}
            </div>
        )
    }

    private onQueryChange = (query: string) => this.setState({ query })

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(this.props.history, { query: this.getQueryPrefix() + this.state.query })
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
}
