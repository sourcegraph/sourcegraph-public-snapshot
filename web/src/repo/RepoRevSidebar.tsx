import CloseIcon from '@sourcegraph/icons/lib/Close'
import ListIcon from '@sourcegraph/icons/lib/List'
import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { makeRepoURI } from '.'
import { gql, queryGraphQL } from '../backend/graphql'
import { Resizable } from '../components/Resizable'
import { Spacer, Tab, Tabs, TabsWithLocalStorageViewStatePersistence } from '../components/Tabs'
import { fetchSite } from '../site-admin/backend'
import { fileHistorySidebarEnabled } from '../site-admin/configHelpers'
import { eventLogger } from '../tracking/eventLogger'
import { Tree } from '../tree/Tree'
import { createAggregateError } from '../util/errors'
import { memoizeObservable } from '../util/memoize'
import { RepoRevSidebarCommits } from './RepoRevSidebarCommits'
import { RepoRevSidebarSymbols } from './RepoRevSidebarSymbols'

const fetchTree = memoizeObservable(
    (args: { repoPath: string; commitID: string }): Observable<string[]> =>
        queryGraphQL(
            gql`
                query FileTree($repoPath: String!, $commitID: String!) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            tree(recursive: true) {
                                files {
                                    path
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
                    !data.repository ||
                    !data.repository.commit ||
                    !data.repository.commit.tree ||
                    !data.repository.commit.tree.files
                ) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.tree.files.map(file => file.path)
            })
        ),
    makeRepoURI
)

type SidebarTabID = 'files' | 'symbols' | 'commits'

interface Props {
    repoID: GQLID
    repoPath: string
    rev: string | undefined
    commitID: string
    filePath: string
    isDir: boolean
    defaultBranch: string
    className: string
    history: H.History
    location: H.Location
}

interface State {
    loading: boolean
    error?: string

    showSidebar: boolean

    /**
     * Should the "File History" sidebar be shown. Set by feature flag.
     */
    showHistorySidebar: boolean

    /**
     * All file paths in the repository.
     */
    files?: string[]
}

/**
 * The sidebar for a specific repo revision that shows the list of files and directories.
 */
export class RepoRevSidebar extends React.PureComponent<Props, State> {
    private static LAST_TAB_STORAGE_KEY = 'repo-rev-sidebar-last-tab'
    private static HIDDEN_STORAGE_KEY = 'repo-rev-sidebar-hidden'

    private static TABS: Tab<SidebarTabID>[] = [{ id: 'files', label: 'Files' }, { id: 'symbols', label: 'Symbols' }]
    private static FILE_TABS: Tab<SidebarTabID>[] = [{ id: 'commits', label: 'File history' }]

    public state: State = {
        loading: true,
        showSidebar: localStorage.getItem(RepoRevSidebar.HIDDEN_STORAGE_KEY) === null,
        showHistorySidebar: false,
    }

    private specChanges = new Subject<{ repoPath: string; commitID: string }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Fetch site config
        this.subscriptions.add(
            fetchSite({ telemetrySamples: false }).subscribe(
                site =>
                    this.setState({
                        showHistorySidebar: fileHistorySidebarEnabled(site.configuration.effectiveContents),
                    }),
                error => this.setState({ error: error.message })
            )
        )

        // Fetch repository revision.
        this.subscriptions.add(
            this.specChanges
                .pipe(switchMap(({ repoPath, commitID }) => fetchTree({ repoPath, commitID })))
                .subscribe(files => this.setState({ files }), err => this.setState({ error: err.message }))
        )
        this.specChanges.next({ repoPath: this.props.repoPath, commitID: this.props.commitID })
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.repoPath !== this.props.repoPath || props.commitID !== this.props.commitID) {
            this.specChanges.next({ repoPath: props.repoPath, commitID: props.commitID })
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.showSidebar) {
            return (
                <button
                    type="button"
                    className={`btn btn-icon repo-rev-sidebar-toggle ${this.props.className}-toggle`}
                    onClick={this.onSidebarToggle}
                    data-tooltip="Show sidebar"
                >
                    <ListIcon />
                </button>
            )
        }

        let tabs = RepoRevSidebar.TABS
        if (!this.props.isDir && this.state.showHistorySidebar) {
            tabs = tabs.concat(RepoRevSidebar.FILE_TABS)
        }

        return (
            <Resizable
                className="repo-rev-container__sidebar-resizable"
                handlePosition="right"
                storageKey="repo-rev-sidebar"
                defaultSize={256 /* px */}
                element={
                    <TabsWithLocalStorageViewStatePersistence
                        tabs={tabs}
                        storageKey={RepoRevSidebar.LAST_TAB_STORAGE_KEY}
                        tabBarEndFragment={
                            <>
                                <Spacer />
                                <button
                                    onClick={this.onSidebarToggle}
                                    className={`btn btn-icon repo-rev-sidebar__close-button ${Tabs.tabBorderClassName}`}
                                    data-tooltip="Close"
                                >
                                    <CloseIcon />
                                </button>
                            </>
                        }
                        id="explorer"
                        className={`repo-rev-sidebar ${this.props.className} ${
                            this.state.showSidebar ? `repo-rev-sidebar--open ${this.props.className}--open` : ''
                        }`}
                        tabClassName="repo-rev-sidebar__tab"
                        onSelectTab={this.onSelectTab}
                    >
                        {this.state.files && (
                            <Tree
                                key="files"
                                repoPath={this.props.repoPath}
                                rev={this.props.rev}
                                history={this.props.history}
                                scrollRootSelector="#explorer"
                                selectedPath={this.props.filePath || ''}
                                paths={this.state.files}
                            />
                        )}
                        {
                            <RepoRevSidebarSymbols
                                key="symbols"
                                repoID={this.props.repoID}
                                rev={this.props.rev}
                                history={this.props.history}
                                location={this.props.location}
                            />
                        }
                        {this.props.isDir ? (
                            undefined
                        ) : (
                            <RepoRevSidebarCommits
                                key="commits"
                                repoID={this.props.repoID}
                                rev={this.props.rev}
                                filePath={this.props.filePath}
                                history={this.props.history}
                                location={this.props.location}
                            />
                        )}
                    </TabsWithLocalStorageViewStatePersistence>
                }
            />
        )
    }

    private onSidebarToggle = () => {
        if (this.state.showSidebar) {
            localStorage.setItem(RepoRevSidebar.HIDDEN_STORAGE_KEY, 'true')
        } else {
            localStorage.removeItem(RepoRevSidebar.HIDDEN_STORAGE_KEY)
        }
        this.setState(state => ({ showSidebar: !state.showSidebar }))
    }

    private onSelectTab = (tab: string) => {
        if (tab === 'symbols') {
            eventLogger.log('SidebarSymbolsTabSelected')
        } else if (tab === 'files') {
            eventLogger.log('SidebarFilesTabSelected')
        } else if (tab === 'commits') {
            eventLogger.log('SidebarCommitsTabSelected')
        }
    }
}
