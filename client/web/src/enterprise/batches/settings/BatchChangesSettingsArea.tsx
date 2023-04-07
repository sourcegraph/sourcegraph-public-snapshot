import React from 'react'

import { mdiPulse } from '@mdi/js'
import { upperFirst } from 'lodash'

import { PageHeader, Text, H3, Container, Icon } from '@sourcegraph/wildcard'
import { BatchChangeRolloutWindow } from '@sourcegraph/shared/src/schema/site.schema'

import { PageTitle } from '../../../components/PageTitle'
import { UserAreaUserFields } from '../../../graphql-operations'

import { UserCodeHostConnections } from './CodeHostConnections'

import styles from './BatchChangesSettingsArea.module.scss'

export interface BatchChangesSettingsAreaProps {
    user: Pick<UserAreaUserFields, 'id'>
}

/** The page area for all batch changes settings. It's shown in the user settings sidebar. */
export const BatchChangesSettingsArea: React.FunctionComponent<
    React.PropsWithChildren<BatchChangesSettingsAreaProps>
> = props => {
    console.log(window.context)
    const { batchChangesRolloutWindows } = window.context;
    return (
        <div className="test-batches-settings-page">
            <PageTitle title="Batch changes settings" />
            <PageHeader headingElement="h2" path={[{ text: 'Batch Changes settings' }]} className="mb-3" />
            {batchChangesRolloutWindows && batchChangesRolloutWindows.length > 0 && <RolloutWindowsConfiguration rolloutWindows={batchChangesRolloutWindows} />}
            <UserCodeHostConnections
                headerLine={<Text>Add access tokens to enable Batch Changes changeset creation on your code hosts.</Text>}
                userID={props.user.id}
            />
        </div>
    )
}

interface RolloutWindowsConfigurationProps {
    rolloutWindows: BatchChangeRolloutWindow[]
}

export const RolloutWindowsConfiguration: React.FunctionComponent<React.PropsWithChildren<RolloutWindowsConfigurationProps>> = ({ rolloutWindows }) => (
    <Container className="mb-3">
        <H3>Rollout Windows</H3>
        <Text>Specifies specific windows, which can have associated rate limits, to be used when reconciling published changesets.</Text>
        <ul className={styles.rolloutWindowList}>
            {rolloutWindows.map((rolloutWindow, index) => (
                <li key={index} className={styles.rolloutWindowListItem}>
                    <Text className={styles.rolloutWindowListItemFrequency}>
                        <Icon className={styles.rolloutWindowListItemFrequencyIcon} svgPath={mdiPulse} aria-label="Copy snippet" />
                        {(typeof rolloutWindow.rate === 'string') ? upperFirst(rolloutWindow.rate.replace('/', ' changesets per ')) : `${rolloutWindow.rate} changesets per minute`}
                    </Text>
                    <small><strong>Days</strong>: {(rolloutWindow.days && rolloutWindow.days.length > 0) ? rolloutWindow.days.join(", ") : 'every other day'}</small>
                    <br />
                    {(rolloutWindow.start && rolloutWindow.end && <small>{rolloutWindow.start} - {rolloutWindow.end} UTC</small>)}
                </li>
            ))}
        </ul>
    </Container>
)
