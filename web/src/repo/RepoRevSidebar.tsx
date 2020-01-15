import * as H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import { FormatListBulletedIcon } from '../../../shared/src/components/icons'
import { Resizable } from '../../../shared/src/components/Resizable'
import {
    Spacer,
    Tab,
    TabBorderClassName,
    TabsWithLocalStorageViewStatePersistence,
} from '../../../shared/src/components/Tabs'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { AbsoluteRepoFile } from '../../../shared/src/util/url'
import { eventLogger } from '../tracking/eventLogger'
import { Tree } from '../tree/Tree'
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

    private static TABS: Tab<SidebarTabID>[] = [
        { id: 'files', label: 'Files' },
        { id: 'symbols', label: 'Symbols' },
    ]

    public state: State = {
        showSidebar: localStorage.getItem(RepoRevSidebar.HIDDEN_STORAGE_KEY) === null,
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Toggle sidebar visibility when the user presses 'alt+s'.
        that.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.altKey && event.key === 's'))
                .subscribe(event => {
                    event.preventDefault()
                    that.setState(prevState => ({ showSidebar: !prevState.showSidebar }))
                })
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!that.state.showSidebar) {
            return (
                <button
                    type="button"
                    className={`btn btn-icon repo-rev-sidebar-toggle ${that.props.className}-toggle`}
                    onClick={that.onSidebarToggle}
                    data-tooltip="Show sidebar (Alt+S/Opt+S)"
                >
                    <FormatListBulletedIcon />
                </button>
            )
        }

        const STORAGE_KEY = 'repo-rev-sidebar'

        return (
            <Resizable
                className="repo-rev-container__sidebar-resizable"
                handlePosition="right"
                storageKey={STORAGE_KEY}
                defaultSize={256 /* px */}
                element={
                    <TabsWithLocalStorageViewStatePersistence
                        tabs={RepoRevSidebar.TABS}
                        storageKey={RepoRevSidebar.LAST_TAB_STORAGE_KEY}
                        tabBarEndFragment={
                            <>
                                <Spacer />
                                <button
                                    type="button"
                                    onClick={that.onSidebarToggle}
                                    className={`btn btn-icon tab_bar__close-button ${TabBorderClassName}`}
                                    title="Close sidebar (Alt+S/Opt+S)"
                                >
                                    <CloseIcon className="icon-inline" />
                                </button>
                            </>
                        }
                        id="explorer"
                        className={`repo-rev-sidebar ${that.props.className} ${
                            that.state.showSidebar ? `repo-rev-sidebar--open ${that.props.className}--open` : ''
                        } e2e-repo-rev-sidebar`}
                        tabClassName="tab-bar__tab--h5like"
                        onSelectTab={that.onSelectTab}
                    >
                        <Tree
                            key="files"
                            repoName={that.props.repoName}
                            rev={that.props.rev}
                            commitID={that.props.commitID}
                            history={that.props.history}
                            location={that.props.location}
                            scrollRootSelector="#explorer"
                            activePath={that.props.filePath}
                            activePathIsDir={that.props.isDir}
                            sizeKey={`Resizable:${STORAGE_KEY}`}
                        />
                        <RepoRevSidebarSymbols
                            key="symbols"
                            repoID={that.props.repoID}
                            rev={that.props.rev}
                            activePath={that.props.filePath}
                            history={that.props.history}
                            location={that.props.location}
                        />
                    </TabsWithLocalStorageViewStatePersistence>
                }
            />
        )
    }

    private onSidebarToggle = (): void => {
        if (that.state.showSidebar) {
            localStorage.setItem(RepoRevSidebar.HIDDEN_STORAGE_KEY, 'true')
        } else {
            localStorage.removeItem(RepoRevSidebar.HIDDEN_STORAGE_KEY)
        }
        that.setState(state => ({ showSidebar: !state.showSidebar }))
    }

    private onSelectTab = (tab: string): void => {
        if (tab === 'symbols') {
            eventLogger.log('SidebarSymbolsTabSelected')
        } else if (tab === 'files') {
            eventLogger.log('SidebarFilesTabSelected')
        } else if (tab === 'history') {
            eventLogger.log('SidebarHistoryTabSelected')
        }
    }
}
