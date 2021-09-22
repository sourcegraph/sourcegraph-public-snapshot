import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import { ConfigurationEditor } from './ConfigurationEditor'
import { RepositoryPolicies } from './RepositoryPolicies'

export interface RepositoryConfigurationProps extends ThemeProps, TelemetryProps {
    repo: { id: string }
    indexingEnabled: boolean
    history: H.History
}

export const RepositoryConfiguration: FunctionComponent<RepositoryConfigurationProps> = ({
    repo,
    indexingEnabled,
    history,
    ...props
}) => (
    <Tabs size="medium">
        <TabList>
            <Tab>Repository-specific policies</Tab>
            <Tab>Global policies</Tab>
            {indexingEnabled && <Tab>Index configuration</Tab>}
        </TabList>

        <TabPanels>
            <TabPanel>
                <RepositoryPolicies isGlobal={false} repo={repo} indexingEnabled={indexingEnabled} history={history} />
            </TabPanel>

            <TabPanel>
                <RepositoryPolicies isGlobal={true} repo={repo} indexingEnabled={indexingEnabled} history={history} />
            </TabPanel>

            {indexingEnabled && (
                <TabPanel>
                    <Container>
                        <h3>Auto-indexing configuration</h3>

                        <ConfigurationEditor repoId={repo.id} history={history} {...props} />
                    </Container>
                </TabPanel>
            )}
        </TabPanels>
    </Tabs>
)
