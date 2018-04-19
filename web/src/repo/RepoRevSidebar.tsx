import CloseIcon from '@sourcegraph/icons/lib/Close'
import ListIcon from '@sourcegraph/icons/lib/List'
import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { makeRepoURI } from '.'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { Resizable } from '../components/Resizable'
import { Spacer, Tab, TabBorderClassName, TabsWithLocalStorageViewStatePersistence } from '../components/Tabs'
import { eventLogger } from '../tracking/eventLogger'
import { Tree } from '../tree/Tree'
import { Tree2 } from '../tree/Tree2'
import { createAggregateError } from '../util/errors'
import { memoizeObservable } from '../util/memoize'
import { RepoRevSidebarSymbols } from './RepoRevSidebarSymbols'

const fetchTree = memoizeObservable(
    (args: { repoPath: string; commitID: string }): Observable<string[]> =>
        queryGraphQL(
            gql`
                query FileTree($repoPath: String!, $commitID: String!) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            tree(recursive: true) {
                                internalRaw
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
                    typeof data.repository.commit.tree.internalRaw !== 'string'
                ) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.tree.internalRaw.split('\0')
            })
        ),
    makeRepoURI
)

type SidebarTabID = 'files' | 'symbols' | 'commits'

interface Props {
    repoID: GQL.ID
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
     * All file paths in the repository.
     */
    files?: string[]
}

// Run `localStorage.oldTree=true;location.reload()` to enable the old file tree in case of issues.
const TreeOrTree2 = localStorage.getItem('oldTree') !== null ? Tree : Tree2

/**
 * The sidebar for a specific repo revision that shows the list of files and directories.
 */
export class RepoRevSidebar extends React.PureComponent<Props, State> {
    private static LAST_TAB_STORAGE_KEY = 'repo-rev-sidebar-last-tab'
    private static HIDDEN_STORAGE_KEY = 'repo-rev-sidebar-hidden'

    private static TABS: Tab<SidebarTabID>[] = [{ id: 'files', label: 'Files' }, { id: 'symbols', label: 'Symbols' }]

    public state: State = {
        loading: true,
        showSidebar: localStorage.getItem(RepoRevSidebar.HIDDEN_STORAGE_KEY) === null,
    }

    private specChanges = new Subject<{ repoPath: string; commitID: string }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Fetch repository revision.
        this.subscriptions.add(
            this.specChanges
                .pipe(switchMap(({ repoPath, commitID }) => fetchTree({ repoPath, commitID })))
                .subscribe(files => this.setState({ files }), err => this.setState({ error: err.message }))
        )
        this.specChanges.next({ repoPath: this.props.repoPath, commitID: this.props.commitID })

        // Toggle sidebar visibility when the user presses 'alt+s'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.altKey && event.keyCode === 83))
                .subscribe(event => {
                    event.preventDefault()
                    this.setState(prevState => ({ showSidebar: !prevState.showSidebar }))
                })
        )
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
                    data-tooltip="Show sidebar (Alt+S/Opt+S)"
                >
                    <ListIcon />
                </button>
            )
        }

        return (
            <Resizable
                className="repo-rev-container__sidebar-resizable"
                handlePosition="right"
                storageKey="repo-rev-sidebar"
                defaultSize={256 /* px */}
                element={
                    <TabsWithLocalStorageViewStatePersistence
                        tabs={RepoRevSidebar.TABS}
                        storageKey={RepoRevSidebar.LAST_TAB_STORAGE_KEY}
                        tabBarEndFragment={
                            <>
                                <Spacer />
                                <button
                                    onClick={this.onSidebarToggle}
                                    className={`btn btn-icon tab_bar__close-button ${TabBorderClassName}`}
                                    data-tooltip="Close sidebar (Alt+S/Opt+S)"
                                >
                                    <CloseIcon />
                                </button>
                            </>
                        }
                        id="explorer"
                        className={`repo-rev-sidebar ${this.props.className} ${
                            this.state.showSidebar ? `repo-rev-sidebar--open ${this.props.className}--open` : ''
                        }`}
                        tabClassName="tab-bar__tab--h5like"
                        onSelectTab={this.onSelectTab}
                    >
                        {this.state.files && (
                            <TreeOrTree2
                                key="files"
                                repoPath={this.props.repoPath}
                                rev={this.props.rev}
                                history={this.props.history}
                                scrollRootSelector="#explorer"
                                activePath={this.props.filePath}
                                selectedPath={this.props.filePath}
                                activePathIsDir={this.props.isDir}
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
