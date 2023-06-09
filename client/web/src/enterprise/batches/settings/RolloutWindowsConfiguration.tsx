import React from 'react'

import { mdiPulse } from '@mdi/js'

import { Text, H3, Container, Icon, LoadingSpinner, ErrorAlert, Link, Code } from '@sourcegraph/wildcard'

import { useBatchChangesRolloutWindowConfig } from '../backend'

import { formatRate, formatDays } from './format'

import styles from './RolloutWindowsConfiguration.module.scss'

// Displays the rollout window configuration.
export const RolloutWindowsConfiguration: React.FunctionComponent = () => {
    const { loading, error, rolloutWindowConfig } = useBatchChangesRolloutWindowConfig()
    return (
        <Container className="mb-3">
            <H3>Rollout windows</H3>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert error={error} />}
            {!loading &&
                rolloutWindowConfig &&
                (rolloutWindowConfig.length === 0 ? (
                    <Text className="mb-0">
                        No rollout windows configured for changesets. Learn how to configure them in{' '}
                        <Link to="/help/admin/config/batch_changes#rollout-windows" target="_blank">
                            site settings.
                        </Link>
                    </Text>
                ) : (
                    <>
                        <Text>
                            Configuring rollout windows allows changesets to be reconciled at a slower or faster rate
                            based on the time of day and/or the day of the week. These windows are applied to changesets
                            across all code hosts and can be configured with the{' '}
                            <Code>batchChanges.rolloutWindows</Code>{' '}
                            <Link to="/help/admin/config/batch_changes#rollout-windows">
                                site configuration option.
                            </Link>
                        </Text>
                        <ul className={styles.rolloutWindowList}>
                            {rolloutWindowConfig.map((rolloutWindow, index) => (
                                <li key={index} className={styles.rolloutWindowListItem}>
                                    <Text className={styles.rolloutWindowListItemFrequency}>
                                        <Icon
                                            className={styles.rolloutWindowListItemFrequencyIcon}
                                            svgPath={mdiPulse}
                                            aria-label="Rollout window frequency"
                                        />
                                        {formatRate(rolloutWindow.rate)}
                                    </Text>
                                    <small>on {formatDays(rolloutWindow.days)}</small>
                                    <br />
                                    {rolloutWindow.start && rolloutWindow.end && (
                                        <small>
                                            {rolloutWindow.start} - {rolloutWindow.end} UTC
                                        </small>
                                    )}
                                </li>
                            ))}
                        </ul>
                    </>
                ))}
        </Container>
    )
}
