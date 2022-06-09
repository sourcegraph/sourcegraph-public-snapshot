import React, { useCallback } from 'react'

import classNames from 'classnames'
import MenuUpIcon from 'mdi-react/MenuUpIcon'

import {
    Button,
    Card,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
    useLocalStorage,
    Icon,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
} from '@sourcegraph/wildcard'

import { PlatformContextProps } from '../../platform/context'
import { ExtensionsControllerProps } from '../controller'

import { ActiveExtensionsPanel } from './ActiveExtensionsPanel'

import styles from './index.module.scss'

export interface ExtensionsDevelopmentToolsProps
    extends ExtensionsControllerProps,
        PlatformContextProps<'sideloadedExtensionURL' | 'settings'> {
    link: React.ComponentType<React.PropsWithChildren<{ id: string }>>
}

const LAST_TAB_STORAGE_KEY = 'ExtensionDevTools.lastTab'

type ExtensionDevelopmentToolsTabID = 'activeExtensions' | 'loggers'

interface ExtensionDevelopmentToolsTab {
    id: ExtensionDevelopmentToolsTabID
    label: string
    component: React.ComponentType<React.PropsWithChildren<ExtensionsDevelopmentToolsProps>>
}

const TABS: ExtensionDevelopmentToolsTab[] = [
    { id: 'activeExtensions', label: 'Active extensions', component: ActiveExtensionsPanel },
]

const ExtensionDevelopmentTools: React.FunctionComponent<
    React.PropsWithChildren<ExtensionsDevelopmentToolsProps>
> = props => {
    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    return (
        <Tabs
            as={Card}
            defaultIndex={tabIndex}
            className={classNames('border-0 rounded-0', styles.extensionStatus)}
            onChange={handleTabsChange}
        >
            <TabList>
                {TABS.map(({ label, id }) => (
                    <Tab key={id} data-tab-content={id}>
                        {label}
                    </Tab>
                ))}
            </TabList>

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
    <Popover>
        <PopoverTrigger
            as={Button}
            className="text-decoration-none px-2"
            variant="link"
            aria-label="Open extensions developer tools"
        >
            <span className="text-muted">Ext</span> <Icon role="img" as={MenuUpIcon} aria-hidden={true} />
        </PopoverTrigger>
        <PopoverContent position={Position.leftEnd}>
            <ExtensionDevelopmentTools {...props} />
        </PopoverContent>
    </Popover>
))
