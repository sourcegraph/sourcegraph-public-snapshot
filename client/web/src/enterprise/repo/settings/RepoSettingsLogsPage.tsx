import { type FC, useCallback, useEffect, useState } from 'react'

import { mdiCalendar, mdiClockAlertOutline } from '@mdi/js'
import { parseISO } from 'date-fns'
import { useSearchParams } from 'react-router-dom'

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
import type { SettingsAreaRepositoryFields, RepositoryRecordedCommandFields } from '../../../graphql-operations'
import { LogsPageTabs } from '../../../repo/constants'
import { eventLogger } from '../../../tracking/eventLogger'

import { useFetchRecordedCommands } from './backend'
import { formatDuration } from './utils'

import styles from './RepoSettingsLogsPage.module.scss'

export interface RepoSettingsLogsPageProps {
    repo: SettingsAreaRepositoryFields
}

/**
 * The repository settings log page.
 */
export const RepoSettingsLogsPage: FC<RepoSettingsLogsPageProps> = ({ repo }) => {
    const [activeTabIndex, setActiveTabIndex] = useState<number>(LogsPageTabs.COMMANDS)
    useEffect(() => eventLogger.logPageView('RepoSettingsLogs'), [])
    const [searchParams, setSearchParams] = useSearchParams()
    const activeTab = searchParams.get('activeTab') ?? LogsPageTabs.COMMANDS.toString()

    const setActiveTab = useCallback(
        (tabIndex: number): void => {
            setActiveTabIndex(tabIndex)
            searchParams.set('activeTab', tabIndex.toString())
            // We set `replace` to true here because we don't want to get pushing
            // to history when navigating between tabs.
            setSearchParams(searchParams, { replace: true })
        },
        [searchParams, setSearchParams]
    )

    useEffect(() => {
        const numericTabIdx = parseInt(activeTab, 10)
        switch (numericTabIdx) {
            case LogsPageTabs.SYNCLOGS:
                setActiveTab(LogsPageTabs.SYNCLOGS)
                break
            default:
                setActiveTab(LogsPageTabs.COMMANDS)
        }
    }, [setActiveTab, activeTab])

    return (
        <>
            <PageTitle title="Logs" />
            <PageHeader path={[{ text: 'Logs and activities' }]} headingElement="h2" className="mb-3" />

            <Container>
                <Tabs
                    size="medium"
                    lazy={true}
                    className={styles.tabContainer}
                    index={activeTabIndex}
                    onChange={setActiveTab}
                >
                    <TabList>
                        <Tab>Last executed commands</Tab>
                        <Tab>Last sync output</Tab>
                    </TabList>

                    <TabPanels>
                        <TabPanel>
                            <LastExecutedCommands repo={repo} />
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

interface LastExecutedCommandsProps {
    repo: SettingsAreaRepositoryFields
}

const LastExecutedCommands: FC<LastExecutedCommandsProps> = ({ repo }) => {
    const { recordedCommands, loading, error, fetchMore, hasNextPage } = useFetchRecordedCommands(repo.id)

    return (
        <>
            {error && <ErrorAlert error={error} />}
            <div aria-label="recorded commands">
                {!loading && recordedCommands.length === 0 && <Text className="my-2">No recorded commands yet.</Text>}
                {recordedCommands.map((command, index) => (
                    // We use the index as key here because commands don't have the concept
                    // of IDs and there's nothing really unique about each command.
                    //
                    // eslint-disable-next-line react/no-array-index-key
                    <LastExecutedCommandNode mirrorInfo={repo.mirrorInfo} command={command} key={index} />
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

interface LastExecutedCommandNodeProps {
    command: RepositoryRecordedCommandFields
    mirrorInfo: SettingsAreaRepositoryFields['mirrorInfo']
}

const LastExecutedCommandNode: FC<LastExecutedCommandNodeProps> = ({ command, mirrorInfo }) => {
    const startDate = parseISO(command.start)
    const duration = formatDuration(command.duration)

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
        'Last sync command did not produce any output'
    return (
        <div className="mt-2">
            <Text>Output from this repository's most recent sync</Text>
            <LogOutput text={output} logDescription="Job output:" />
        </div>
    )
}
