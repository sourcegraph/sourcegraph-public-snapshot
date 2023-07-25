import { FC, useEffect, useState } from 'react'

import { mdiChevronUp, mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'
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
    Button,
    Collapse,
    CollapseHeader,
    CollapsePanel,
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
        } else {
            setActiveTab(LogsPageTabs.COMMANDS)
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
                            <LastGitCommands recordedCommands={repo.recordedCommands} />
                        </TabPanel>

                        <TabPanel>
                            <LastSyncOutput repo={repo} />
                        </TabPanel>
                    </TabPanels>
                </Tabs>
            </Container>
        </>
    )
}

interface LastGitCommandsProps {
    recordedCommands: SettingsAreaRepositoryFields['recordedCommands']
}

const LastGitCommands: FC<LastGitCommandsProps> = props => {
    const { recordedCommands } = props

    if (recordedCommands.length === 0) {
        return <Text className="my-2">No recorded commands for repository.</Text>
    }

    return (
        <div className="mt-2">
            {recordedCommands.map((command, index) => (
                // We use the index as key here because commands don't have the concept
                // of IDs and there's nothing really unique about each command.
                //
                // eslint-disable-next-line react/no-array-index-key
                <LastGitCommandNode command={command} key={index} name={`Command ${index + 1}`} />
            ))}
        </div>
    )
}

interface LastGitCommandNodeProps {
    command: SettingsAreaRepositoryFields['recordedCommands'][0]
    name: string
}

const LastGitCommandNode: FC<LastGitCommandNodeProps> = ({ command, name }) => {
    const [isOpened, setIsOpened] = useState(false)
    const startDate = new Date(command.start)

    let duration: string
    if (command.duration > 1) {
        duration = `${command.duration.toFixed(2)}s`
    } else {
        const durationInMs = command.duration * 1000
        duration = `${durationInMs.toFixed(2)}ms`
    }

    return (
        <Collapse isOpen={isOpened} onOpenChange={setIsOpened}>
            <CollapseHeader
                as={Button}
                outline={true}
                focusLocked={false}
                variant="secondary"
                className={classNames('w-100 my-2 text-left', styles.commandNode)}
            >
                <Icon aria-hidden={true} svgPath={isOpened ? mdiChevronUp : mdiChevronDown} className="mr-1" />
                <Timestamp date={startDate} />
                <Text className="mb-0">{name}</Text>
                <Text className="mb-0">{duration}</Text>
            </CollapseHeader>
            <CollapsePanel>
                <LogOutput text={command.command} logDescription="Command:" />
            </CollapsePanel>
        </Collapse>
    )
}

interface LastSyncOutputProps {
    repo: SettingsAreaRepositoryFields
}

const LastSyncOutput: FC<LastSyncOutputProps> = props => {
    const output =
        (props.repo.mirrorInfo.cloneInProgress && 'Cloning in progress...') ||
        props.repo.mirrorInfo.lastSyncOutput ||
        'No logs yet.'
    return (
        <div className="mt-2">
            <Text>Output from this repository's most recent sync</Text>
            <LogOutput text={output} logDescription="Job output:" />
        </div>
    )
}
