import classNames from 'classnames'
import * as React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Tabs, Tab, TabList, TabPanel, TabPanels } from '@sourcegraph/wildcard'

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
}

export const HomePanels: React.FunctionComponent<Props> = (props: Props) => (
    <div className={classNames('container', styles.homePanels)} data-testid="home-panels">
        <div className="row">
            <RepositoriesPanel {...props} className={classNames('col-lg-4', styles.panel)} />
            <RecentSearchesPanel {...props} className={classNames('col-lg-8', styles.panel)} />
        </div>
        <div className="row">
            <RecentFilesPanel {...props} className={classNames('col-lg-7', styles.panel)} />

            {props.isSourcegraphDotCom ? (
                <div className={classNames('col-lg-5', styles.panel)}>
                    <Tabs defaultIndex={1} className="h-100">
                        <TabList>
                            <Tab>Community search contexts</Tab>
                            <Tab>Invite colleagues</Tab>
                        </TabList>
                        <TabPanels className="h-100">
                            <TabPanel>
                                <CommunitySearchContextsPanel {...props} hideTitle={true} />
                            </TabPanel>
                            <TabPanel className="h-100">
                                <CollaboratorsPanel {...props} />
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
