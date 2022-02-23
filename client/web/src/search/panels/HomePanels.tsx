import classNames from 'classnames'
import * as React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Tabs, Tab, TabList, TabPanel, TabPanels, useLocalStorage } from '@sourcegraph/wildcard'

import { HomePanelsProps } from '..'
import { AuthenticatedUser } from '../../auth'

import { CollaboratorsPanel } from './CollaboratorsPanel'
import { CommunitySearchContextsPanel } from './CommunitySearchContextPanel'
import styles from './HomePanels.module.scss'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'

interface Props extends TelemetryProps, HomePanelsProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    showCollaborators: boolean
}

const INVITES_TAB_KEY = 'HomePage.UserInvites.Tab'

export const HomePanels: React.FunctionComponent<Props> = (props: Props) => {
    const hasEmailSendingCapabilities = true
    const [persistedTabIndex, setPersistedTabIndex] = useLocalStorage(INVITES_TAB_KEY, 1)

    return (
        <div className={classNames('container', styles.homePanels)} data-testid="home-panels">
            <div className="row">
                <RepositoriesPanel {...props} className={classNames('col-lg-4', styles.panel)} />
                <RecentSearchesPanel {...props} className={classNames('col-lg-8', styles.panel)} />
            </div>
            <div className="row">
                <RecentFilesPanel {...props} className={classNames('col-lg-7', styles.panel)} />

                {props.isSourcegraphDotCom ? (
                    props.showCollaborators ? (
                        <div className={classNames('col-lg-5', styles.panel)}>
                            <Tabs defaultIndex={persistedTabIndex} onChange={setPersistedTabIndex} className="h-100">
                                <TabList>
                                    <Tab>Community search contexts</Tab>
                                    <Tab>Invite colleagues</Tab>
                                </TabList>
                                <TabPanels className="h-100">
                                    <TabPanel className="h-100">
                                        <CommunitySearchContextsPanel {...props} hideTitle={true} />
                                    </TabPanel>
                                    <TabPanel className="h-100">
                                        <CollaboratorsPanel
                                            hasEmailSendingCapabilities={hasEmailSendingCapabilities}
                                            {...props}
                                        />
                                    </TabPanel>
                                </TabPanels>
                            </Tabs>
                        </div>
                    ) : (
                        <CommunitySearchContextsPanel {...props} className={classNames('col-lg-5', styles.panel)} />
                    )
                ) : props.showCollaborators ? (
                    <div className={classNames('col-lg-5', styles.panel)}>
                        <Tabs
                            defaultIndex={persistedTabIndex}
                            onChange={setPersistedTabIndex}
                            className={styles.tabPanel}
                        >
                            <TabList>
                                <Tab>Saved searches</Tab>
                                <Tab>Invite colleagues</Tab>
                            </TabList>
                            <TabPanels className="h-100">
                                <TabPanel className="h-100">
                                    <SavedSearchesPanel {...props} hideTitle={true} />
                                </TabPanel>
                                <TabPanel className="h-100">
                                    <CollaboratorsPanel
                                        hasEmailSendingCapabilities={hasEmailSendingCapabilities}
                                        {...props}
                                    />
                                </TabPanel>
                            </TabPanels>
                        </Tabs>
                    </div>
                ) : (
                    <SavedSearchesPanel {...props} className={classNames('col-lg-5', styles.panel)} />
                )}
            </div>
        </div>
    )
}
