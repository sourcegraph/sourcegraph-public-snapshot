import React, { useEffect, useState } from 'react'

import { mdiPulse } from '@mdi/js'
import * as jsonc from 'jsonc-parser'
import { upperFirst, capitalize } from 'lodash'

import { BatchChangeRolloutWindow } from '@sourcegraph/shared/src/schema/site.schema'
import { PageHeader, Text, H3, Container, Icon, LoadingSpinner, ErrorAlert, Link } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'
import { UserAreaUserFields } from '../../../graphql-operations'
import { useGetBatchChangesSiteConfiguration } from '../backend'

import { UserCodeHostConnections } from './CodeHostConnections'

import styles from './BatchChangesSettingsArea.module.scss'

export interface BatchChangesSettingsAreaProps {
    user: Pick<UserAreaUserFields, 'id'>
}

/** The page area for all batch changes settings. It's shown in the user settings sidebar. */
export const BatchChangesSettingsArea: React.FunctionComponent<
    React.PropsWithChildren<BatchChangesSettingsAreaProps>
> = props => (
    <div className="test-batches-settings-page">
        <PageTitle title="Batch changes settings" />
        <PageHeader headingElement="h2" path={[{ text: 'Batch Changes settings' }]} className="mb-3" />
        <RolloutWindowsConfiguration />
        <UserCodeHostConnections
            headerLine={<Text>Add access tokens to enable Batch Changes changeset creation on your code hosts.</Text>}
            userID={props.user.id}
        />
    </div>
)

// Displays the rollout window configuration.
export const RolloutWindowsConfiguration: React.FunctionComponent = () => {
    const [rolloutWindows, setRolloutWindows] = useState<BatchChangeRolloutWindow[]>([])
    const { loading, error, data } = useGetBatchChangesSiteConfiguration()
    useEffect(() => {
        if (data) {
            const siteConfig = jsonc.parse(data.site.configuration.effectiveContents)
            console.log(siteConfig['batchChanges.rolloutWindows'])
            setRolloutWindows(siteConfig['batchChanges.rolloutWindows'] || [])
        }
    }, [data])
    return (
        <Container className="mb-3">
            <H3>Rollout Windows</H3>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert error={error} />}
            {!loading &&
                data &&
                (rolloutWindows.length === 0 ? (
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
                            <strong>batchChanges.rolloutWindows</strong>{' '}
                            <Link to="/help/admin/config/batch_changes#rollout-windows">
                                site configuration option.
                            </Link>
                        </Text>
                        <ul className={styles.rolloutWindowList}>
                            {rolloutWindows.map((rolloutWindow, index) => (
                                <li key={index} className={styles.rolloutWindowListItem}>
                                    <Text className={styles.rolloutWindowListItemFrequency}>
                                        <Icon
                                            className={styles.rolloutWindowListItemFrequencyIcon}
                                            svgPath={mdiPulse}
                                            aria-label="Rollout window frequency"
                                        />
                                        {formatRate(rolloutWindow.rate)}
                                    </Text>
                                    <small>on: {formatDays(rolloutWindow.days)}</small>
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

/**
 * Formats the days of the week for a rollout window for display.
 *
 * If days are provided, joins them with commas and capitalizes each day name.
 * Otherwise returns 'every other day' as the default.
 *
 * @param days The days of the week for the rollout window, e.g. ['monday', 'wednesday']
 * @returns The formatted days for display in the UI
 */
const formatDays = (days: string[] | undefined): string => {
    if (days && Array.isArray(days) && days.length > 0) {
        return days.join(', ').replace(/\w+/g, capitalize)
    }

    return 'every other day'
}

/**
 * Formats the rollout window rate for display.
 *
 * According to the schema, if the rate is a number then it can only be zero.
 * If the rate starts with '0/' then we revert to displaying None, since this is the same as 0.
 * Otherwise, we display the rate in a readable format, e.g. '2 changesets per minute'.
 *
 * https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@3ee30bb/-/blob/schema/site.schema.json?L567-571
 *
 * @param rate The rollout window rate, either a number or a string like '1/minute'
 * @returns The formatted rate for display
 */
const formatRate = (rate: string | number): string => {
    if (typeof rate === 'number' || rate.startsWith('0/')) {
        return 'None'
    }
    return upperFirst(rate.replace('/', ' changesets per '))
}
