import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import * as H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { FormatListBulletedIcon } from '../../../shared/src/components/icons'
import { Resizable } from '../../../shared/src/components/Resizable'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { Scalars } from '../../../shared/src/graphql-operations'
import { ThemeProps } from '../../../shared/src/theme'
import { AbsoluteRepoFile } from '../../../shared/src/util/url'
import { Tree } from '../tree/Tree'
import { useLocalStorage } from '../util/useLocalStorage'
import { RepoRevisionSidebarSymbols } from './RepoRevisionSidebarSymbols'

function useMultiKeyPress(): any {
    const [keysPressed, setKeyPressed] = useState(new Set([]))

    useEffect(() => {
        function downHandler({ key }: { key: string[] }): void {
            setKeyPressed(keysPressed.add(key))
        }

        function upHandler({ key }: { key: string[] }): void {
            keysPressed.delete(key)
            setKeyPressed(keysPressed)
        }
        window.addEventListener('keydown', downHandler)
        window.addEventListener('keyup', upHandler)
        return () => {
            window.removeEventListener('keydown', downHandler)
            window.removeEventListener('keyup', upHandler)
        }
    }, [keysPressed]) // Empty array ensures that effect is only run on mount and unmount

    return keysPressed
}

interface Props extends AbsoluteRepoFile, ExtensionsControllerProps, ThemeProps {
    repoID: Scalars['ID']
    isDir: boolean
    defaultBranch: string
    className: string
    history: H.History
    location: H.Location
    repoName: any
    revision: any
    commitID: any
    filePath: any
    extensionsController: any
    isLightTheme: any
}

/**
 * The sidebar for a specific repo revision that shows the list of files and directories.
 */

function areKeysPressed(keys: string[], keysPressed: Set<string>): boolean {
    const required = new Set(keys)
    for (const element of keysPressed) {
        required.delete(element)
    }
    return required.size === 0
}

export const RepoRevisionSidebar: React.FunctionComponent<Props> = props => {
    // public componentDidMount(): void {
    //     // Toggle sidebar visibility when the user presses 'alt+s'.
    //     this.subscriptions.add(
    //         fromEvent<KeyboardEvent>(window, 'keydown')
    //             .pipe(filter(event => event.altKey && event.key === 's'))
    //             .subscribe(event => {
    //                 event.preventDefault()
    //                 this.setState(previousState => ({ showSidebar: !previousState.showSidebar }))
    //             })
    //     )
    // }

    const keysPressed = useMultiKeyPress()
    const hsrfPressed = areKeysPressed(['s'], keysPressed)

    console.log(hsrfPressed)

    const [tabIndex, setTabIndex] = useLocalStorage('repo-revision-sidebar-last-tab', 0)
    const [showSidebar, setShowSidebar] = useLocalStorage('repo-revision-sidebar-hidden', false)

    const STORAGE_KEY = 'repo-revision-sidebar'

    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])
    const sidebarToggle = useCallback(() => setShowSidebar(!showSidebar), [setShowSidebar, showSidebar])

    if (showSidebar) {
        return (
            <button
                type="button"
                className="btn btn-icon repo-revision-sidebar-toggle repo-revision-container__sidebar-toggle"
                onClick={sidebarToggle}
                data-tooltip="Show sidebar (Alt+S/Opt+S)"
            >
                <FormatListBulletedIcon className="icon-inline" />
            </button>
        )
    }

    return (
        <Resizable
            className="repo-revision-container__sidebar-resizable"
            handlePosition="right"
            storageKey={STORAGE_KEY}
            defaultSize={256 /* px */}
            element={
                <Tabs className="w-100" defaultIndex={tabIndex} onChange={handleTabsChange}>
                    <TabList>
                        <Tab>Files</Tab>
                        <Tab>Symbols</Tab>
                    </TabList>
                    <div className="d-flex overflow-auto" style={{ height: '88vh' }}>
                        <TabPanels className="w-100">
                            <TabPanel>
                                <Tree
                                    key="files"
                                    repoName={props.repoName}
                                    revision={props.revision}
                                    commitID={props.commitID}
                                    history={props.history}
                                    location={props.location}
                                    scrollRootSelector="#explorer"
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
            }
        />
    )
}
