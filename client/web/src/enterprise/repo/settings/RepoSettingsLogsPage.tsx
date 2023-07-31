import { FC, useEffect, useState } from 'react'

import { mdiCalendar, mdiClockAlertOutline } from '@mdi/js'
import { parseISO } from 'date-fns'
import { useLocation } from 'react-router-dom'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import {
    Text,
    PageHeader,
    Container,
    Tabs,
    Tab,
    TabList,
    TabPanels,
    TabPanel,
    Icon,
    LoadingSpinner,
    ErrorAlert,
    Button,
} from '@sourcegraph/wildcard'

import { LogOutput } from '../../../components/LogOutput'
import { PageTitle } from '../../../components/PageTitle'
import { SettingsAreaRepositoryFields, RepositoryRecordedCommandFields } from '../../../graphql-operations'
import { LogsPageTabs } from '../../../repo/constants'
import { eventLogger } from '../../../tracking/eventLogger'

import { useFetchRecordedCommands } from './backend'

import styles from './RepoSettingsLogsPage.module.scss'

export interface RepoSettingsLogsPageProps {
    repo: SettingsAreaRepositoryFields
}

/**
 * The repository settings log page.
 */
export const RepoSettingsLogsPage: FC<RepoSettingsLogsPageProps> = ({ repo }) => {
    const [activeTab, setActiveTab] = useState<number>(LogsPageTabs.COMMANDS)
    useEffect(() => eventLogger.logPageView('RepoSettingsLogs'), [])
    const location = useLocation()

    useEffect(() => {
        const searchParams = new URLSearchParams(location.search)
        const tab = searchParams.get('activeTab')
        if (tab) {
            const activeTabIdx = parseInt(tab, 10)
            switch (activeTabIdx) {
                case LogsPageTabs.SYNCLOGS:
                    setActiveTab(LogsPageTabs.SYNCLOGS)
                    break
                case LogsPageTabs.COMMANDS:
                default:
                    setActiveTab(LogsPageTabs.COMMANDS)
            }
        }
    }, [location.search])

    const setActiveTabIndex = (index: number): void => setActiveTab(index)

    return (
        <>
            <PageTitle title="Logs" />
            <PageHeader path={[{ text: 'Logs and activities' }]} headingElement="h2" className="mb-3" />

            <Container>
                <Tabs
                    size="medium"
                    lazy={true}
                    className={styles.tabContainer}
                    index={activeTab}
                    onChange={setActiveTabIndex}
                >
                    <TabList>
                        <Tab>Last repo commands</Tab>
                        <Tab>Last sync output</Tab>
                    </TabList>

                    <TabPanels>
                        <TabPanel>
                            <LastRepoCommands repo={repo} />
                        </TabPanel>

                        <TabPanel>
                            <LastSyncOutput mirrorInfo={repo.mirrorInfo} />
                        </TabPanel>
                    </TabPanels>
                </Tabs>
            </Container>
        </>
    )
}

interface LastRepoCommandsProps {
    repo: SettingsAreaRepositoryFields
}

const LastRepoCommands: FC<LastRepoCommandsProps> = ({ repo }) => {
    const { recordedCommands, loading, error, fetchMore, hasNextPage } = useFetchRecordedCommands(repo.id)

    return (
        <>
            {error && <ErrorAlert error={error} />}
            <div aria-label="recorded commands">
                {(!loading && recordedCommands.length === 0) && <Text className="my-2">No recorded commands yet.</Text>}
                {recordedCommands.map((command, index) => (
                    // We use the index as key here because commands don't have the concept
                    // of IDs and there's nothing really unique about each command.
                    //
                    // eslint-disable-next-line react/no-array-index-key
                    <LastRepoCommandNode mirrorInfo={repo.mirrorInfo} command={command} key={index} />
                ))}
            </div>
            {loading && <LoadingSpinner />}
            {hasNextPage && !loading && (
                <div className="d-flex justify-content-center">
                    <Button onClick={() => fetchMore(recordedCommands.length)}>Show more</Button>
                </div>
            )}
        </>
    )
}

interface LastRepoCommandNodeProps {
    command: RepositoryRecordedCommandFields
    mirrorInfo: SettingsAreaRepositoryFields['mirrorInfo']
}

const LastRepoCommandNode: FC<LastRepoCommandNodeProps> = ({ command, mirrorInfo }) => {
    const startDate = parseISO(command.start)

    let duration: string
    if (command.duration > 1) {
        duration = `${command.duration.toFixed(2)}s`
    } else {
        const durationInMs = command.duration * 1000
        duration = `${durationInMs.toFixed(2)}ms`
    }

    return (
        <div className={styles.commandNode}>
            <div className={styles.commandNodeHeader}>
                <div>
                    <Icon aria-hidden={true} svgPath={mdiCalendar} className="mr-1" />
                    <Timestamp date={startDate} />
                </div>

                {/* Replace this we type when we have the type field */}
                <Text />

                <div className={styles.commandNodeDurationGroup}>
                    <Icon aria-hidden={true} svgPath={mdiClockAlertOutline} className="mr-1" />
                    <Text className="mb-0">Ran in {duration}</Text>
                </div>
            </div>

            <LogOutput className={styles.commandNodeLogOutput} text={command.command} logDescription="Command:" />

            <div className={styles.commandNodeFooter}>
                <small className="mt-2 mb-0">
                    <span className="font-weight-bold">Shard:</span> {mirrorInfo.shard}
                </small>
                {command.dir && (
                    <small className="mt-2 mb-0">
                        <span className="font-weight-bold">Path:</span> {command.dir}
                    </small>
                )}
            </div>
        </div>
    )
}

interface LastSyncOutputProps {
    mirrorInfo: SettingsAreaRepositoryFields['mirrorInfo']
}

const LastSyncOutput: FC<LastSyncOutputProps> = props => {
    const output =
        (props.mirrorInfo.cloneInProgress && 'Cloning in progress...') ||
        props.mirrorInfo.lastSyncOutput ||
        'No logs yet.'
    return (
        <div className="mt-2">
            <Text>Output from this repository's most recent sync</Text>
            <LogOutput text={output} logDescription="Job output:" />
        </div>
    )
}
