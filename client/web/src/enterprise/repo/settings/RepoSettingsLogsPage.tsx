import { FC, useEffect, useState } from 'react'

import { mdiChevronUp, mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'
import { formatDistanceToNowStrict } from 'date-fns'

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
import { eventLogger } from '../../../tracking/eventLogger'

import styles from './RepoSettingsLogsPage.module.scss'

export interface RepoSettingsLogsPageProps {
    repo: SettingsAreaRepositoryFields
}

/**
 * The repository settings log page.
 */
export const RepoSettingsLogsPage: FC<RepoSettingsLogsPageProps> = ({ repo }) => {
    useEffect(() => eventLogger.logPageView('RepoSettingsLogs'))

    return (
        <>
            <PageTitle title="Logs and activities" />
            <PageHeader path={[{ text: 'Logs and activities' }]} headingElement="h2" className="mb-3" />

            <Container>
                <div className="form-group">
                    <Input value={repo.name} readOnly={true} className="mb-0" />
                </div>

                <Tabs size="medium" lazy={true} className={styles.tabContainer}>
                    <TabList>
                        <Tab>Last Git commands</Tab>
                        <Tab>Last Git logs</Tab>
                    </TabList>

                    <TabPanels>
                        <TabPanel>
                            <LastGitCommands recordedCommands={repo.recordedCommands} />
                        </TabPanel>

                        <TabPanel>
                            <LastRepoGitLogs repo={repo} />
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

    return (
        <div className="mt-2">
            {recordedCommands.map((command, index) => (
                // We use the index as key here because commands dont have the concept
                // of IDs and there's nothing really unique about each command.
                //
                // eslint-disable-next-line react/no-array-index-key
                <LastGitCommandNode command={command} key={index} />
            ))}
        </div>
    )
}

interface LastGitCommandNodeProps {
    command: SettingsAreaRepositoryFields['recordedCommands'][0]
}

const LastGitCommandNode: FC<LastGitCommandNodeProps> = ({ command }) => {
    const [isOpened, setIsOpened] = useState(false)
    const startDate = new Date(command.start)
    return (
        <Collapse isOpen={isOpened} onOpenChange={setIsOpened}>
            <CollapseHeader
                as={Button}
                outline={true}
                focusLocked={true}
                variant="secondary"
                className={classNames('w-100 my-2 text-left', styles.lastGitCommandNode)}
            >
                <Icon aria-hidden={true} svgPath={isOpened ? mdiChevronUp : mdiChevronDown} className="mr-1" />
                <Text className="mb-0">{formatDistanceToNowStrict(startDate)}</Text>
                <Text className="mb-0">Janitor Run</Text>
                <Text className="mb-0">{command.duration.toFixed(2)}s</Text>
            </CollapseHeader>
            <CollapsePanel>
                <LogOutput text={command.command} logDescription="Command output:" />
                <Button variant="primary" className="mt-2">
                    Rerun
                </Button>
            </CollapsePanel>
        </Collapse>
    )
}

interface LastRepoGitLogsProps {
    repo: SettingsAreaRepositoryFields
}

const LastRepoGitLogs: FC<LastRepoGitLogsProps> = props => {
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
