import { FC, useEffect, useState } from 'react'

import { mdiCalendar, mdiClockAlertOutline } from '@mdi/js'
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
    Input,
    Icon,
} from '@sourcegraph/wildcard'

import { LogOutput } from '../../../components/LogOutput'
import { PageTitle } from '../../../components/PageTitle'
import { SettingsAreaRepositoryFields } from '../../../graphql-operations'
import { LogsPageTabs } from '../../../repo/constants'
import { eventLogger } from '../../../tracking/eventLogger'

import styles from './RepoSettingsLogsPage.module.scss'

export interface RepoSettingsLogsPageProps {
    repo: SettingsAreaRepositoryFields
}

/**
 * The repository settings log page.
 */
export const RepoSettingsLogsPage: FC<RepoSettingsLogsPageProps> = ({ repo }) => {
    const [activeTab, setActiveTab] = useState<number>(LogsPageTabs.COMMANDS)
    useEffect(() => eventLogger.logPageView('RepoSettingsLogs'))
    const location = useLocation()

    useEffect(() => {
        const searchParams = new URLSearchParams(location.search)
        if (searchParams.has('activeTab')) {
            switch (searchParams.get('activeTab')) {
                case LogsPageTabs.SYNCLOGS.toString():
                    setActiveTab(LogsPageTabs.SYNCLOGS)
                    break
                case LogsPageTabs.COMMANDS.toString():
                default:
                    setActiveTab(LogsPageTabs.COMMANDS)
            }
        }
    }, [location.search])

    const handleActiveTab = (index: number): void => setActiveTab(index)

    return (
        <>
            <PageTitle title="Logs" />
            <PageHeader path={[{ text: 'Logs and activities' }]} headingElement="h2" className="mb-3" />

            <Container>
                <div className="form-group">
                    <Input value={repo.name} readOnly={true} className="mb-0" />
                </div>

                <Tabs
                    size="medium"
                    lazy={true}
                    className={styles.tabContainer}
                    index={activeTab}
                    onChange={handleActiveTab}
                >
                    <TabList>
                        <Tab>Last repo commands</Tab>
                        <Tab>Last sync output</Tab>
                    </TabList>

                    <TabPanels>
                        <TabPanel>
                            <LastRepoCommands
                                mirrorInfo={repo.mirrorInfo}
                                recordedCommands={repo.recordedCommands.slice(0, 5)}
                            />
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
    recordedCommands: SettingsAreaRepositoryFields['recordedCommands']
    mirrorInfo: SettingsAreaRepositoryFields['mirrorInfo']
}

const LastRepoCommands: FC<LastRepoCommandsProps> = ({ recordedCommands, mirrorInfo }) => {
    if (recordedCommands.length === 0) {
        return <Text className="my-3">No recorded commands for repository.</Text>
    }

    return (
        <div className="mt-2">
            {recordedCommands.map((command, index) => (
                // We use the index as key here because commands don't have the concept
                // of IDs and there's nothing really unique about each command.
                //
                // eslint-disable-next-line react/no-array-index-key
                <LastRepoCommandNode mirrorInfo={mirrorInfo} command={command} key={index} />
            ))}
        </div>
    )
}

interface LastRepoCommandNodeProps {
    command: SettingsAreaRepositoryFields['recordedCommands'][0]
    mirrorInfo: SettingsAreaRepositoryFields['mirrorInfo']
}

const LastRepoCommandNode: FC<LastRepoCommandNodeProps> = ({ command, mirrorInfo }) => {
    const startDate = new Date(command.start)

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
                <small className="mt-2 mb-0">
                    <span className="font-weight-bold">Path:</span> {command.dir}
                </small>
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
