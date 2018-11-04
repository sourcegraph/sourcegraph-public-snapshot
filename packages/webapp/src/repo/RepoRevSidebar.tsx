import * as H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import { AbsoluteRepoFile } from '.'
import * as GQL from '../backend/graphqlschema'
import { Resizable } from '../components/Resizable'
import { Spacer, Tab, TabBorderClassName, TabsWithLocalStorageViewStatePersistence } from '../components/Tabs'
import { ExtensionsControllerProps } from '../extensions/ExtensionsClientCommonContext'
import { eventLogger } from '../tracking/eventLogger'
import { Tree } from '../tree/Tree'
import { FormatListBulletedIcon } from '../util/icons' // TODO: Switch to mdi icon
import { RepoRevSidebarSymbols } from './RepoRevSidebarSymbols'

type SidebarTabID = 'files' | 'symbols' | 'history'

interface Props extends AbsoluteRepoFile, ExtensionsControllerProps {
    repoID: GQL.ID
    isDir: boolean
    defaultBranch: string
    className: string
    history: H.History
    location: H.Location
}

interface State {
    showSidebar: boolean
}

/**
 * The sidebar for a specific repo revision that shows the list of files and directories.
 */
export class RepoRevSidebar extends React.PureComponent<Props, State> {
    private static LAST_TAB_STORAGE_KEY = 'repo-rev-sidebar-last-tab'
    private static HIDDEN_STORAGE_KEY = 'repo-rev-sidebar-hidden'

    private static TABS: Tab<SidebarTabID>[] = [{ id: 'files', label: 'Files' }, { id: 'symbols', label: 'Symbols' }]

    public state: State = {
        showSidebar: localStorage.getItem(RepoRevSidebar.HIDDEN_STORAGE_KEY) === null,
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Toggle sidebar visibility when the user presses 'alt+s'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.altKey && event.key === 's'))
                .subscribe(event => {
                    event.preventDefault()
                    this.setState(prevState => ({ showSidebar: !prevState.showSidebar }))
                })
        )
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
                    <FormatListBulletedIcon />
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
                                    title="Close sidebar (Alt+S/Opt+S)"
                                >
                                    <CloseIcon className="icon-inline" />
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
                        <Tree
                            key="files"
                            repoPath={this.props.repoPath}
                            rev={this.props.rev}
                            commitID={this.props.commitID}
                            history={this.props.history}
                            location={this.props.location}
                            scrollRootSelector="#explorer"
                            activePath={this.props.filePath}
                            activePathIsDir={this.props.isDir}
                        />
                        <RepoRevSidebarSymbols
                            key="symbols"
                            repoID={this.props.repoID}
                            rev={this.props.rev}
                            history={this.props.history}
                            location={this.props.location}
                        />
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
        } else if (tab === 'history') {
            eventLogger.log('SidebarHistoryTabSelected')
        }
    }
}
