import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import {
    getConfigurationForRepository as defaultGetConfigurationForRepository,
    getInferredConfigurationForRepository as defaultGetInferredConfigurationForRepository,
    updateConfigurationForRepository as defaultUpdateConfigurationForRepository,
} from './backend'
import { ConfigurationEditor } from './ConfigurationEditor'
import { GlobalPolicies } from './GlobalPolicies'
import { RepositoryPolicies } from './RepositoryPolicies'

export interface RepositoryConfigurationProps extends ThemeProps, TelemetryProps {
    repo: { id: string }
    disabled: boolean
    deleting: boolean
    policies?: CodeIntelligenceConfigurationPolicyFields[]
    deletePolicy: (id: string, name: string) => Promise<void>
    globalPolicies?: CodeIntelligenceConfigurationPolicyFields[]
    deleteGlobalPolicy: (id: string, name: string) => Promise<void>
    deleteError?: Error
    updateConfigurationForRepository: typeof defaultUpdateConfigurationForRepository
    getConfigurationForRepository: typeof defaultGetConfigurationForRepository
    getInferredConfigurationForRepository: typeof defaultGetInferredConfigurationForRepository
    indexingEnabled: boolean
    history: H.History
}

export const RepositoryConfiguration: FunctionComponent<RepositoryConfigurationProps> = ({
    repo,
    disabled,
    deleting,
    policies,
    deletePolicy,
    globalPolicies,
    deleteGlobalPolicy,
    deleteError,
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
                <RepositoryPolicies
                    disabled={disabled}
                    deleting={deleting}
                    policies={policies}
                    deletePolicy={deletePolicy}
                    indexingEnabled={indexingEnabled}
                    history={history}
                />
            </TabPanel>

            <TabPanel>
                <GlobalPolicies
                    repo={repo}
                    disabled={disabled}
                    deleting={deleting}
                    globalPolicies={globalPolicies}
                    deleteGlobalPolicy={deleteGlobalPolicy}
                    indexingEnabled={indexingEnabled}
                    history={history}
                />
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
