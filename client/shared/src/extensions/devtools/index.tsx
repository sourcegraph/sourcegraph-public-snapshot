import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback } from 'react'
import { UncontrolledPopover } from 'reactstrap'

import { PlatformContextProps } from '../../platform/context'
import { useLocalStorage } from '../../util/useLocalStorage'
import { ExtensionsControllerProps } from '../controller'

import { ActiveExtensionsPanel } from './ActiveExtensionsPanel'

export interface ExtensionsDevelopmentToolsProps
    extends ExtensionsControllerProps,
        PlatformContextProps<'sideloadedExtensionURL' | 'settings'> {
    link: React.ComponentType<{ id: string }>
}

const LAST_TAB_STORAGE_KEY = 'ExtensionDevTools.lastTab'

type ExtensionDevelopmentToolsTabID = 'activeExtensions' | 'loggers'

interface ExtensionDevelopmentToolsTab {
    id: ExtensionDevelopmentToolsTabID
    label: string
    component: React.ComponentType<ExtensionsDevelopmentToolsProps>
}

const TABS: ExtensionDevelopmentToolsTab[] = [
    { id: 'activeExtensions', label: 'Active extensions', component: ActiveExtensionsPanel },
]

const ExtensionDevelopmentTools: React.FunctionComponent<ExtensionsDevelopmentToolsProps> = props => {
    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    return (
        <Tabs defaultIndex={tabIndex} className="extension-status card border-0 rounded-0" onChange={handleTabsChange}>
            <div className="tablist-wrapper w-100 align-items-center">
                <TabList>
                    {TABS.map(({ label, id }) => (
                        <Tab className="d-flex flex-1 justify-content-around" key={id} data-tab-content={id}>
                            {label}
                        </Tab>
                    ))}
                </TabList>
            </div>

            <TabPanels>
                {TABS.map(tab => (
                    <TabPanel key={tab.id}>
                        <tab.component {...props} />
                    </TabPanel>
                ))}
            </TabPanels>
        </Tabs>
    )
}

/** A button that toggles the visibility of the ExtensionDevTools element in a popover. */
export const ExtensionDevelopmentToolsPopover = React.memo<ExtensionsDevelopmentToolsProps>(props => (
    <>
        <button type="button" id="extension-status-popover" className="btn btn-link text-decoration-none px-2">
            <span className="text-muted">Ext</span> <MenuUpIcon className="icon-inline" />
        </button>
        <UncontrolledPopover
            placement="auto-end"
            target="extension-status-popover"
            hideArrow={true}
            popperClassName="border-0 rounded-0"
        >
            <ExtensionDevelopmentTools {...props} />
        </UncontrolledPopover>
    </>
))
