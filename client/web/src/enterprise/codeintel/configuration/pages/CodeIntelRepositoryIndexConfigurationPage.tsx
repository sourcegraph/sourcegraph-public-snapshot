import { FunctionComponent, useCallback, useEffect, useState } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Link, Tabs, TabList, Tab, TabPanels, TabPanel } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { CodeIntelConfigurationPageHeader } from '../components/CodeIntelConfigurationPageHeader'
import { ConfigurationEditor } from '../components/ConfigurationEditor'
import { ConfigurationForm } from '../components/ConfigurationForm'

export interface CodeIntelRepositoryIndexConfigurationPageProps extends TelemetryProps {
    repo: { id: string }
    authenticatedUser: AuthenticatedUser | null
}

export const CodeIntelRepositoryIndexConfigurationPage: FunctionComponent<
    CodeIntelRepositoryIndexConfigurationPageProps
> = ({ repo, authenticatedUser, telemetryService, ...props }) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelRepositoryIndexConfiguration'), [telemetryService])

    const [activeTabIndex, setActiveTabIndex] = useState<number>(0)
    const setTab = useCallback((index: number) => {
        setActiveTabIndex(index)
    }, [])

    return (
        <>
            <PageTitle title="Code graph data repository index configuration" />
            <CodeIntelConfigurationPageHeader>
                <PageHeader
                    headingElement="h2"
                    path={[
                        {
                            text: <>Code graph data repository index configuration</>,
                        },
                    ]}
                    description={
                        <>
                            Provide explicit index job configuration to customize how this repository is indexed. See
                            the{' '}
                            <Link to="/help/code_navigation/references/auto_indexing_configuration">
                                reference guide
                            </Link>{' '}
                            for more information.
                        </>
                    }
                    className="mb-3"
                />
            </CodeIntelConfigurationPageHeader>
            <Tabs size="large" index={activeTabIndex} onChange={setTab} lazy={true}>
                <TabList>
                    <Tab key="form">Form</Tab>
                    <Tab key="raw">Raw</Tab>
                </TabList>
                <TabPanels>
                    <TabPanel>
                        <ConfigurationForm
                            repoId={repo.id}
                            authenticatedUser={authenticatedUser}
                            telemetryService={telemetryService}
                        />
                    </TabPanel>
                    <TabPanel>
                        <ConfigurationEditor
                            repoId={repo.id}
                            authenticatedUser={authenticatedUser}
                            telemetryService={telemetryService}
                            {...props}
                        />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </>
    )
}
