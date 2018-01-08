import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { defer } from 'rxjs/observable/defer'
import { delay } from 'rxjs/operators/delay'
import { map } from 'rxjs/operators/map'
import { retryWhen } from 'rxjs/operators/retryWhen'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { makeRepoURI } from '.'
import { gql, queryGraphQL } from '../backend/graphql'
import { HeroPage } from '../components/HeroPage'
import { ChromeExtensionToast, FirefoxExtensionToast } from '../marketing/BrowserExtensionToast'
import { SurveyToast } from '../marketing/SurveyToast'
import { IS_CHROME, IS_FIREFOX } from '../marketing/util'
import { memoizeObservable } from '../util/memoize'
import { CopyLinkAction } from './actions/CopyLinkAction'
import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import {
    CloneInProgressError,
    ECLONEINPROGESS,
    EREPONOTFOUND,
    EREVNOTFOUND,
    ERREPOSEEOTHER,
    RepoNotFoundError,
    RepoSeeOtherError,
    RevNotFoundError,
} from './backend'
import { BlobPage } from './BlobPage'
import { DirectoryPage } from './DirectoryPage'
import { RepoHeaderActionPortal } from './RepoHeaderActionPortal'
import { RepoRevSidebar } from './RepoRevSidebar'
import { RevSwitcher } from './RevSwitcher'

const fetchRepositoryRevision = memoizeObservable(
    (args: { repoPath: string; rev: string }): Observable<GQL.IRepository | null> =>
        queryGraphQL(
            gql`
                query RepositoryRevision($repoPath: String!, $rev: String!) {
                    repository(uri: $repoPath) {
                        uri
                        commit(rev: $rev) {
                            cloneInProgress
                            commit {
                                sha1
                            }
                        }
                        defaultBranch
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                if (data.repository && data.repository.redirectURL) {
                    throw new RepoSeeOtherError(data.repository.redirectURL)
                }
                if (!data.repository || !data.repository.commit) {
                    throw new RepoNotFoundError(args.repoPath)
                }
                if (data.repository.commit.cloneInProgress) {
                    throw new CloneInProgressError(args.repoPath)
                }
                if (!data.repository.commit.commit) {
                    throw new RevNotFoundError(args.rev)
                }
                if (!data.repository.defaultBranch) {
                    throw new RevNotFoundError('HEAD')
                }
                return data.repository
            })
        ),
    makeRepoURI
)

interface Props extends RouteComponentProps<{ filePath: string }> {
    repo: GQL.IRepository
    rev: string | undefined
    user: GQL.IUser | null
    objectType: 'blob' | 'tree'
}

interface State {
    loading: boolean
    error?: string

    showSidebar: boolean

    /**
     * The repository object with revision fields populated.
     */
    repo?: GQL.IRepository | null
}

/**
 * A container for a repository page that incorporates revisioned Git data. (For example,
 * blob and tree pages are revisioned, but the repository settings page is not.)
 */
export class RepoRevContainer extends React.PureComponent<Props, State> {
    public state: State = {
        loading: true,
        showSidebar: true,
    }

    private repoRevChanges = new Subject<{ repo: string; rev: string | undefined }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Fetch repository revision.
        this.subscriptions.add(
            this.repoRevChanges
                .pipe(
                    switchMap(({ repo, rev }) =>
                        defer(() => fetchRepositoryRevision({ repoPath: repo, rev: rev || 'HEAD' }))
                            // On a CloneInProgress error, retry after 1s
                            .pipe(
                                retryWhen(errors =>
                                    errors.pipe(
                                        tap(err => {
                                            switch (err.code) {
                                                case ERREPOSEEOTHER:
                                                    const externalHostURL = new URL(
                                                        (err as RepoSeeOtherError).redirectURL
                                                    )
                                                    const redirectURL = new URL(window.location.href)
                                                    // Preserve the path of the current URL and redirect to the repo on the external host.
                                                    redirectURL.host = externalHostURL.host
                                                    redirectURL.port = externalHostURL.port
                                                    redirectURL.protocol = externalHostURL.protocol
                                                    window.location.href = redirectURL.toString()
                                                case ECLONEINPROGESS:
                                                    // Display cloning screen to the user and retry
                                                    this.setState({
                                                        repo: { commit: { cloneInProgress: true } },
                                                    } as any) // TODO!(sqs) hack
                                                    return
                                                case EREPONOTFOUND:
                                                case EREVNOTFOUND:
                                                    // Display 404 to the user and do not retry
                                                    this.setState({ repo: { commit: {} } } as any) // TODO!(sqs) hack
                                            }
                                            // Don't retry
                                            throw err
                                        }),
                                        delay(1000)
                                    )
                                )
                            )
                    )
                )
                .subscribe(repo => this.setState({ repo }), err => this.setState({ error: err.message }))
        )
        this.repoRevChanges.next({ repo: this.props.repo.uri, rev: this.props.rev })
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.repo !== this.props.repo || props.rev !== this.props.rev) {
            this.repoRevChanges.next({ repo: props.repo.uri, rev: props.rev })
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.repo) {
            return null // loading
        }

        if (this.state.repo.commit.cloneInProgress) {
            return (
                <HeroPage
                    icon={RepoIcon}
                    title={this.props.repo.uri
                        .split('/')
                        .slice(1)
                        .join('/')}
                    subtitle="Cloning in progress"
                />
            )
        }
        if (!this.state.repo.commit.commit) {
            return (
                <HeroPage
                    icon={DirectionalSignIcon}
                    title="404: Not Found"
                    subtitle="The requested revision was not found."
                />
            )
        }

        return (
            <div className="repo-rev-container">
                {IS_CHROME && <ChromeExtensionToast />}
                {IS_FIREFOX && <FirefoxExtensionToast />}
                <SurveyToast />
                <RepoHeaderActionPortal
                    position="right"
                    component={<CopyLinkAction key="copy-link" location={this.props.location} />}
                />
                <RepoHeaderActionPortal
                    position="left"
                    component={
                        <RevSwitcher
                            key="rev-switcher"
                            repoPath={this.props.repo.uri}
                            rev={this.props.rev || this.state.repo.defaultBranch || 'HEAD'}
                            history={this.props.history}
                        />
                    }
                />
                <RepoHeaderActionPortal
                    position="left"
                    component={
                        <GoToPermalinkAction
                            key="go-to-permalink"
                            rev={this.props.rev}
                            commitID={this.state.repo.commit.commit.sha1}
                            location={this.props.location}
                            history={this.props.history}
                        />
                    }
                />
                <RepoRevSidebar
                    className="repo-rev-container__sidebar"
                    repoPath={this.props.repo.uri}
                    rev={this.props.rev}
                    commitID={this.state.repo.commit.commit.sha1}
                    filePath={this.props.match.params.filePath || ''}
                    defaultBranch={this.state.repo.defaultBranch || 'HEAD'}
                    history={this.props.history}
                />
                <div className="repo-rev-container__content">
                    {this.props.objectType === 'tree' && (
                        <DirectoryPage
                            repoPath={this.props.repo.uri}
                            commitID={this.state.repo.commit.commit.sha1}
                            rev={this.props.rev}
                            filePath={this.props.match.params.filePath || ''}
                        />
                    )}
                    {this.props.objectType === 'blob' && (
                        <BlobPage
                            repoPath={this.props.repo.uri}
                            commitID={this.state.repo.commit.commit.sha1}
                            rev={this.props.rev}
                            filePath={this.props.match.params.filePath || ''}
                            location={this.props.location}
                            history={this.props.history}
                        />
                    )}
                </div>
            </div>
        )
    }
}
