import { type FC, useCallback, useEffect, useState } from 'react'

import { mdiAlertCircleOutline, mdiCheckCircleOutline, mdiClockOutline, mdiChevronDown, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'
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
    Alert,
    Link,
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
    useEffect(() => {
        window.context.telemetryRecorder?.recordEvent('repoSettingsLogs', 'viewed')
        eventLogger.logPageView('RepoSettingsLogs')
    }, [window.context.telemetryRecorder])
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
                        <Tab>Command logs</Tab>
                        <Tab>Sync output</Tab>
                    </TabList>

                    <TabPanels>
                        <TabPanel>
                            <CommandLogs repo={repo} />
                        </TabPanel>

                        <TabPanel>
                            <SyncOutput mirrorInfo={repo.mirrorInfo} />
                        </TabPanel>
                    </TabPanels>
                </Tabs>
            </Container>
        </>
    )
}

interface CommandLogsProps {
    repo: SettingsAreaRepositoryFields
}

const CommandLogs: FC<CommandLogsProps> = ({ repo }) => {
    const { recordedCommands, loading, error, fetchMore, hasNextPage, isRecordingEnabled } = useFetchRecordedCommands(
        repo.id
    )
    return (
        <>
            {error && <ErrorAlert error={error} />}
            <div aria-label="recorded commands">
                {/*
                    We explicitly check is `isRecordingEnabled` is false because when fetching this field
                    from the API, `isRecordingEnabled` will be undefined and we don't want to display this
                    instruction until we're certain the repository isn't configured for recording.
                 */}
                {!loading && isRecordingEnabled === false && (
                    <Alert variant="info" className="mt-3">
                        <small>Command recording isn't enabled for this repository.</small>{' '}
                        <small>
                            Visit <Link to="/help/admin/repo/recording">the docs</Link> to learn how to enable command
                            recording.
                        </small>
                    </Alert>
                )}
                {!loading && recordedCommands.length === 0 && isRecordingEnabled && (
                    <Text className="my-2">No recorded commands yet.</Text>
                )}
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
    const [isExpanded, setIsExpanded] = useState(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    const startDate = parseISO(command.start)
    const duration = formatDuration(command.duration)

    return (
        <div className={styles.commandNode}>
            <div className={styles.commandNodeHeader}>
                <div>
                    <Icon
                        aria-hidden={true}
                        svgPath={command.isSuccess ? mdiCheckCircleOutline : mdiAlertCircleOutline}
                        className={classNames(styles.commandNodeStatus, {
                            [styles.commandNodeSuccessStatus]: command.isSuccess,
                            [styles.commandNodeErrorStatus]: !command.isSuccess,
                        })}
                    />
                    <span className="font-weight-bold">{command.isSuccess ? 'Succeeded' : 'Failed'}</span>{' '}
                    <span className="text-muted">
                        <Timestamp date={startDate} /> on shard {mirrorInfo.shard}
                    </span>
                </div>

                <div className={styles.commandNodeDurationGroup}>
                    <Icon aria-hidden={true} svgPath={mdiClockOutline} className="mr-1 text-muted" />
                    <Text className="mb-0 text-muted">{duration}</Text>
                </div>
            </div>

            <LogOutput
                className={classNames(styles.commandNodeLogOutput, {
                    [styles.commandNodeLogOutputFailState]: !command.isSuccess,
                })}
                text={command.command}
                logDescription="Command:"
            />

            <div className={styles.commandNodeFooter}>
                <div>
                    {command.output && (
                        <Button
                            variant="icon"
                            aria-label={isExpanded ? 'Hide output' : 'Show output'}
                            onClick={toggleIsExpanded}
                            className="mr-2"
                        >
                            <Icon
                                aria-hidden={true}
                                svgPath={isExpanded ? mdiChevronDown : mdiChevronRight}
                                className={styles.commandNodeExpandBtn}
                            />
                            <small>Command output</small>
                        </Button>
                    )}
                </div>
                <small className={classNames('text-muted', styles.commandNodePath)}>Path: {command.dir}</small>
                {isExpanded && (
                    <LogOutput text={command.output} logDescription="Output:" className={styles.commandNodeOutput} />
                )}
            </div>
        </div>
    )
}

interface SyncOutputProps {
    mirrorInfo: SettingsAreaRepositoryFields['mirrorInfo']
}

const SyncOutput: FC<SyncOutputProps> = props => {
    const output =
        (props.mirrorInfo.cloneInProgress && 'Cloning in progress...') ||
        props.mirrorInfo.lastSyncOutput ||
        'Last sync command did not produce any output'
    return (
        <div className="mt-2">
            <Text className="mb-1">Output from this repository's most recent sync</Text>
            <LogOutput text={output} logDescription="Job output:" />
        </div>
    )
}
