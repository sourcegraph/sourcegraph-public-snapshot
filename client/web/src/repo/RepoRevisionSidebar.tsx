import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import * as H from 'history'
import React, { useCallback } from 'react'
import { Button } from 'reactstrap'
import { FormatListBulletedIcon } from '../../../shared/src/components/icons'
import { Resizable } from '../../../shared/src/components/Resizable'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { Scalars } from '../../../shared/src/graphql-operations'
import { ThemeProps } from '../../../shared/src/theme'
import { AbsoluteRepoFile } from '../../../shared/src/util/url'
import { Tree } from '../tree/Tree'
import { useLocalStorage } from '../util/useLocalStorage'
import { RepoRevisionSidebarSymbols } from './RepoRevisionSidebarSymbols'

interface Props extends AbsoluteRepoFile, ExtensionsControllerProps, ThemeProps {
    repoID: Scalars['ID']
    isDir: boolean
    defaultBranch: string
    className: string
    history: H.History
    location: H.Location
}

/**
 * The sidebar for a specific repo revision that shows the list of files and directories.
 */
export const RepoRevisionSidebar: React.FunctionComponent<Props> = props => {
    // TODO: make (Alt+S/Opt+S) keyboard logic on panel
    const STORAGE_KEY = 'repo-revision-sidebar'
    const TABS_KEY = 'repo-revision-sidebar-last-tab'
    const SIDEBAR_KEY = 'repo-revision-sidebar-toggle'

    const [tabIndex, setTabIndex] = useLocalStorage(TABS_KEY, 0)
    const [toggleSidebar, setToggleSidebar] = useLocalStorage(SIDEBAR_KEY, true)

    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])
    const handleSidebarToggle = useCallback(() => setToggleSidebar(!toggleSidebar), [setToggleSidebar, toggleSidebar])

    if (!toggleSidebar) {
        return (
            <button
                type="button"
                className="btn btn-icon repo-revision-sidebar-toggle repo-revision-container__sidebar-toggle"
                onClick={handleSidebarToggle}
                data-tooltip="Show sidebar (Alt+S/Opt+S)"
            >
                <FormatListBulletedIcon className="icon-inline" />
            </button>
        )
    }

    return (
        // eslint-disable-next-line react/forbid-dom-props
        <Resizable defaultSize={256} position="left">
            <Tabs className="w-100" defaultIndex={tabIndex} onChange={handleTabsChange}>
                <div className="d-flex">
                    <TabList>
                        <Tab>Files</Tab>
                        <Tab>Symbols</Tab>
                    </TabList>
                    <Button
                        onClick={handleSidebarToggle}
                        close={true}
                        className="bg-transparent border-0 close ml-auto"
                        title="Close sidebar (Alt+S/Opt+S)"
                    />
                </div>
                <div aria-hidden={true} className="d-flex overflow-auto repo-revision-container__tabpanels explorer">
                    <TabPanels className="w-100">
                        <TabPanel tabIndex={-1}>
                            <Tree
                                key="files"
                                repoName={props.repoName}
                                revision={props.revision}
                                commitID={props.commitID}
                                history={props.history}
                                location={props.location}
                                scrollRootSelector=".explorer"
                                activePath={props.filePath}
                                activePathIsDir={props.isDir}
                                sizeKey={`Resizable:${STORAGE_KEY}`}
                                extensionsController={props.extensionsController}
                                isLightTheme={props.isLightTheme}
                            />
                        </TabPanel>
                        <TabPanel>
                            <RepoRevisionSidebarSymbols
                                key="symbols"
                                repoID={props.repoID}
                                revision={props.revision}
                                activePath={props.filePath}
                                history={props.history}
                                location={props.location}
                            />
                        </TabPanel>
                    </TabPanels>
                </div>
            </Tabs>
        </Resizable>
    )
}
