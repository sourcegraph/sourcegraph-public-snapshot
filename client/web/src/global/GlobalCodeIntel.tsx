import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import classNames from 'classnames'
import * as H from 'history'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback } from 'react'
import { UncontrolledPopover } from 'reactstrap'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Button, useLocalStorage } from '@sourcegraph/wildcard'

import styles from './GlobalCodeIntel.module.scss'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    location: H.Location
    onWebHoverOverlayClick: (url: string) => void
}

const SHOW_COOL_CODEINTEL = localStorage.getItem('coolCodeIntel') !== null

export const GlobalCodeIntel: React.FunctionComponent<Props> = props =>
    SHOW_COOL_CODEINTEL ? (
        <ul className={classNames('nav', styles.globalCodeintel)}>
            <li className="nav-item">
                <CoolCodeIntelPopover
                    extensionsController={props.extensionsController}
                    platformContext={props.platformContext}
                />
            </li>
        </ul>
    ) : null

export interface CoolCodeIntelPopoverProps
    extends ExtensionsControllerProps,
        PlatformContextProps<'sideloadedExtensionURL' | 'settings'> {}

const LAST_TAB_STORAGE_KEY = 'ExtensionDevTools.lastTab'

type ExtensionDevelopmentToolsTabID = 'references'

interface CoolCodeIntelToolsTab {
    id: ExtensionDevelopmentToolsTabID
    label: string
    component: React.ComponentType<CoolCodeIntelPopoverProps>
}

export const ReferencesPanel: React.FunctionComponent<CoolCodeIntelPopoverProps> = props => (
    <>
        <div className="card-header">References</div>
        <div className="card-body border-top">
            <h4>
                Check out this <i>intelligence</i>
            </h4>
        </div>
    </>
)
const TABS: CoolCodeIntelToolsTab[] = [{ id: 'references', label: 'References', component: ReferencesPanel }]

const CoolCodeIntel: React.FunctionComponent<CoolCodeIntelPopoverProps> = props => {
    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    return (
        <Tabs
            defaultIndex={tabIndex}
            className={classNames('card border-0 rounded-0', styles.coolCodeIntelStatus)}
            onChange={handleTabsChange}
        >
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
export const CoolCodeIntelPopover = React.memo<CoolCodeIntelPopoverProps>(props => (
    <>
        <Button id="extension-status-popover" className="text-decoration-none px-2" variant="link">
            <span className="text-muted">Cool Code Intel</span> <MenuUpIcon className="icon-inline" />
        </Button>
        <UncontrolledPopover
            placement="bottom"
            target="extension-status-popover"
            hideArrow={true}
            popperClassName="border-0 rounded-0"
        >
            <CoolCodeIntel {...props} />
        </UncontrolledPopover>
    </>
))
