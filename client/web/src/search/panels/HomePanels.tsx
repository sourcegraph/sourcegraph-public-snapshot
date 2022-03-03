import classNames from 'classnames'
import * as React from 'react'
import { useEffect } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Tabs, Tab, TabList, TabPanel, TabPanels } from '@sourcegraph/wildcard'

import { HomePanelsProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { eventLogger } from '../../tracking/eventLogger'

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

const INVITES_TAB_KEY = 'homepage.userInvites.tab'

export const HomePanels: React.FunctionComponent<Props> = (props: Props) => {
    useEffect(() => {
        if (props.showCollaborators === true) {
            return
        }
        const loggerPayload = {
            // The other types are emitted in <CollaboratorsPanel />
            type: 'config-disabled',
        }
        eventLogger.log('HomepageInvitationsViewEmpty', loggerPayload, loggerPayload)
    }, [props.showCollaborators])

    return (
        <div className={classNames('container', styles.homePanels)} data-testid="home-panels">
            <div className="row">
                <RepositoriesPanel {...props} className={classNames('col-lg-4', styles.panel)} />
                <RecentSearchesPanel {...props} className={classNames('col-lg-8', styles.panel)} />
            </div>
            <div className="row">
                <RecentFilesPanel {...props} className={classNames('col-lg-7', styles.panel)} />

                {props.showCollaborators ? (
                    <CollaboratorsTabPanel {...props} />
                ) : props.isSourcegraphDotCom ? (
                    <CommunitySearchContextsPanel {...props} className={classNames('col-lg-5', styles.panel)} />
                ) : (
                    <SavedSearchesPanel {...props} className={classNames('col-lg-5', styles.panel)} />
                )}
            </div>
        </div>
    )
}

const CollaboratorsTabPanel: React.FunctionComponent<Props> = (props: Props) => {
    const [persistedTabIndex, setPersistedTabIndex] = useTemporarySetting(INVITES_TAB_KEY, 1)

    if (persistedTabIndex === undefined) {
        return null
    }

    return (
        <div className={classNames('col-lg-5', styles.panel)}>
            <Tabs defaultIndex={persistedTabIndex} onChange={setPersistedTabIndex} className={styles.tabPanel}>
                <TabList>
                    <Tab>{props.isSourcegraphDotCom ? 'Community search contexts' : 'Saved searches'}</Tab>
                    <Tab>Invite colleagues</Tab>
                </TabList>
                <TabPanels className="h-100">
                    <TabPanel className="h-100">
                        {props.isSourcegraphDotCom ? (
                            <CommunitySearchContextsPanel {...props} insideTabPanel={true} />
                        ) : (
                            <SavedSearchesPanel {...props} insideTabPanel={true} />
                        )}
                    </TabPanel>
                    <TabPanel className="h-100">
                        <CollaboratorsPanel {...props} />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </div>
    )
}
