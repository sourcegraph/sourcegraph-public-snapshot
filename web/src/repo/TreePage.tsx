import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { escapeRegExp, upperFirst } from 'lodash'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import FolderIcon from 'mdi-react/FolderIcon'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import TagIcon from 'mdi-react/TagIcon'
import UserIcon from 'mdi-react/UserIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap, tap } from 'rxjs/operators'
import { ActionItem } from '../../../shared/src/actions/ActionItem'
import { ActionsContainer } from '../../../shared/src/actions/ActionsContainer'
import { ContributableMenu } from '../../../shared/src/api/protocol'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { RepositoryIcon } from '../../../shared/src/components/icons' // TODO: Switch to mdi icon
import { displayRepoName } from '../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { memoizeObservable } from '../../../shared/src/util/memoizeObservable'
import { queryGraphQL } from '../backend/graphql'
import { FilteredConnection } from '../components/FilteredConnection'
import { Form } from '../components/Form'
import { PageTitle } from '../components/PageTitle'
import { isDiscussionsEnabled } from '../discussions'
import { DiscussionsList } from '../discussions/DiscussionsList'
import { searchQueryForRepoRev } from '../search'
import { submitSearch } from '../search/helpers'
import { QueryInput } from '../search/input/QueryInput'
import { SearchButton } from '../search/input/SearchButton'
import { ThemeProps } from '../theme'
import { eventLogger, EventLoggerProps } from '../tracking/eventLogger'
import { basename } from '../util/path'
import { fetchTree } from './backend'
import { GitCommitNode, GitCommitNodeProps } from './commits/GitCommitNode'
import { gitCommitFragment } from './commits/RepositoryCommitsPage'

const TreeEntry: React.FunctionComponent<{
    isDir: boolean
    name: string
    parentPath: string
    url: string
}> = ({ isDir, name, parentPath, url }) => {
    const filePath = parentPath ? parentPath + '/' + name : name
    return (
        <Link to={url} className="tree-entry" title={filePath}>
            {name}
            {isDir && '/'}
        </Link>
    )
}

/**
 * Use a multi-column layout for tree entries when there are at least this many. See TreePage.scss
 * for more information.
 */
const MIN_ENTRIES_FOR_COLUMN_LAYOUT = 6

const TreeEntriesSection: React.FunctionComponent<{
    title: string
    parentPath: string
    entries: Pick<GQL.ITreeEntry, 'name' | 'isDirectory' | 'url'>[]
}> = ({ title, parentPath, entries }) =>
    entries.length > 0 ? (
        <section className="tree-page__section">
            <h3 className="tree-page__section-header">{title}</h3>
            <div className={entries.length > MIN_ENTRIES_FOR_COLUMN_LAYOUT ? 'tree-page__entries--columns' : undefined}>
                {entries.map((e, i) => (
                    <TreeEntry key={i} isDir={e.isDirectory} name={e.name} parentPath={parentPath} url={e.url} />
                ))}
            </div>
        </section>
    ) : null

const fetchTreeCommits = memoizeObservable(
    (args: {
        repo: GQL.ID
        revspec: string
        first?: number
        filePath?: string
    }): Observable<GQL.IGitCommitConnection> =>
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

interface Props
    extends SettingsCascadeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        EventLoggerProps,
        ActivationProps {
    repoName: string
    repoID: GQL.ID
    repoDescription: string
    /** The tree's path in TreePage. We call it filePath for consistency elsewhere. */
    filePath: string
    commitID: string
    rev: string
    location: H.Location
    history: H.History
}

interface State {
    /** This tree, or an error. Undefined while loading. */
    treeOrError?: GQL.IGitTree | ErrorLike

    /**
     * The value of the search query input field.
     */
    query: string
}

export class TreePage extends React.PureComponent<Props, State> {
    public state: State = { query: '' }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    private logViewEvent(props: Props): void {
        if (props.filePath === '') {
            eventLogger.logViewEvent('Repository')
        } else {
            eventLogger.logViewEvent('Tree')
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(
                        (x, y) =>
                            x.repoName === y.repoName &&
                            x.rev === y.rev &&
                            x.commitID === y.commitID &&
                            x.filePath === y.filePath
                    ),
                    tap(props => this.logViewEvent(props)),
                    switchMap(props =>
                        fetchTree({
                            repoName: props.repoName,
                            commitID: props.commitID,
                            rev: props.rev,
                            filePath: props.filePath,
                            first: 2500,
                        }).pipe(
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
        let queryPrefix = searchQueryForRepoRev(this.props.repoName, this.props.rev)
        if (this.props.filePath) {
            queryPrefix += `file:^${escapeRegExp(this.props.filePath)}/ `
        }
        return queryPrefix
    }

    public render(): JSX.Element | null {
        return (
            <div className="tree-page">
                <PageTitle title={this.getPageTitle()} />
                {this.state.treeOrError === undefined && (
                    <div>
                        <LoadingSpinner className="icon-inline tree-page__entries-loader" /> Loading files and
                        directories
                    </div>
                )}
                {this.state.treeOrError !== undefined &&
                    (isErrorLike(this.state.treeOrError) ? (
                        <div className="alert alert-danger">{upperFirst(this.state.treeOrError.message)}</div>
                    ) : (
                        <>
                            {this.state.treeOrError.isRoot ? (
                                <header>
                                    <h2 className="tree-page__title">
                                        <RepositoryIcon className="icon-inline" />{' '}
                                        {displayRepoName(this.props.repoName)}
                                    </h2>
                                    {this.props.repoDescription && <p>{this.props.repoDescription}</p>}
                                    <div className="btn-group mb-3">
                                        <Link
                                            className="btn btn-secondary"
                                            to={`${this.state.treeOrError.url}/-/commits`}
                                        >
                                            <SourceCommitIcon className="icon-inline" /> Commits
                                        </Link>
                                        <Link className="btn btn-secondary" to={`/${this.props.repoName}/-/branches`}>
                                            <SourceBranchIcon className="icon-inline" /> Branches
                                        </Link>
                                        <Link className="btn btn-secondary" to={`/${this.props.repoName}/-/tags`}>
                                            <TagIcon className="icon-inline" /> Tags
                                        </Link>
                                        <Link
                                            className="btn btn-secondary"
                                            to={
                                                this.props.rev
                                                    ? `/${this.props.repoName}/-/compare/...${encodeURIComponent(
                                                          this.props.rev
                                                      )}`
                                                    : `/${this.props.repoName}/-/compare`
                                            }
                                        >
                                            <HistoryIcon className="icon-inline" /> Compare
                                        </Link>
                                        <Link
                                            className={`btn btn-secondary`}
                                            to={`/${this.props.repoName}/-/stats/contributors`}
                                        >
                                            <UserIcon className="icon-inline" /> Contributors
                                        </Link>
                                        {/* TODO(sqs): Inject via enterprise */}
                                        <Link className={`btn btn-secondary`} to={`/${this.props.repoName}/-/checks`}>
                                            <CheckCircleIcon className="icon-inline" /> Checks
                                        </Link>
                                    </div>
                                </header>
                            ) : (
                                <header>
                                    <h2 className="tree-page__title">
                                        <FolderIcon className="icon-inline" /> {this.props.filePath}
                                    </h2>
                                </header>
                            )}
                            <section className="tree-page__section">
                                <h3 className="tree-page__section-header">
                                    Search in this {this.props.filePath ? 'tree' : 'repository'}
                                </h3>
                                <Form className="tree-page__section-search" onSubmit={this.onSubmit}>
                                    <QueryInput
                                        {...this.props}
                                        value={this.state.query}
                                        onChange={this.onQueryChange}
                                        prependQueryForSuggestions={this.getQueryPrefix()}
                                        autoFocus={true}
                                        placeholder=""
                                    />
                                    <SearchButton />
                                </Form>
                            </section>
                            <TreeEntriesSection
                                title="Files and directories"
                                parentPath={this.props.filePath}
                                entries={this.state.treeOrError.entries}
                            />
                            {isDiscussionsEnabled(this.props.settingsCascade) && (
                                <div className="tree-page__section mt-2 tree-page__section--discussions">
                                    <h3 className="tree-page__section-header">Discussions</h3>
                                    <DiscussionsList
                                        {...this.props}
                                        filePath={this.props.filePath + '/**' || undefined}
                                        noun="discussion in this tree"
                                        pluralNoun="discussions in this tree"
                                        defaultFirst={2}
                                        hideSearch={true}
                                        compact={false}
                                    />
                                </div>
                            )}
                            <ActionsContainer
                                {...this.props}
                                menu={ContributableMenu.DirectoryPage}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={items => (
                                    <section className="tree-page__section">
                                        <h3 className="tree-page__section-header">Actions</h3>
                                        {items.map((item, i) => (
                                            <ActionItem
                                                {...this.props}
                                                key={i}
                                                {...item}
                                                className="btn btn-secondary mr-1 mb-1"
                                            />
                                        ))}
                                    </section>
                                )}
                                empty={null}
                            />
                            <div className="tree-page__section">
                                <h3 className="tree-page__section-header">Changes</h3>
                                <FilteredConnection<
                                    GQL.IGitCommit,
                                    Pick<GitCommitNodeProps, 'repoName' | 'className' | 'compact'>
                                >
                                    {...this.props}
                                    className="mt-2 tree-page__section--commits"
                                    listClassName="list-group list-group-flush"
                                    noun="commit in this tree"
                                    pluralNoun="commits in this tree"
                                    queryConnection={this.queryCommits}
                                    nodeComponent={GitCommitNode}
                                    nodeComponentProps={{
                                        repoName: this.props.repoName,
                                        className: 'list-group-item',
                                        compact: true,
                                    }}
                                    updateOnChange={`${this.props.repoName}:${this.props.rev}:${this.props.filePath}`}
                                    defaultFirst={7}
                                    shouldUpdateURLQuery={false}
                                    hideSearch={true}
                                />
                            </div>
                        </>
                    ))}
            </div>
        )
    }

    private onQueryChange = (query: string) => this.setState({ query })

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(
            this.props.history,
            this.getQueryPrefix() + this.state.query,
            this.props.filePath ? 'tree' : 'repo',
            this.props.activation
        )
    }

    private getPageTitle(): string {
        const repoStr = displayRepoName(this.props.repoName)
        if (this.props.filePath) {
            return `${basename(this.props.filePath)} - ${repoStr}`
        }
        return `${repoStr}`
    }

    private queryCommits = (args: { first?: number }) =>
        fetchTreeCommits({
            ...args,
            repo: this.props.repoID,
            revspec: this.props.rev || '',
            filePath: this.props.filePath,
        })
}
