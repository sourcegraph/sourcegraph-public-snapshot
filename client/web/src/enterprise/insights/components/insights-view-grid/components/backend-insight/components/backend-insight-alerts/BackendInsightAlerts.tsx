import React, { FC } from 'react'

import { mdiAlertCircle as mdiAlertCircleOutline } from '@mdi/js'
import classNames from 'classnames'
import { timeFormat } from 'd3-time-format'
import ProgressWrench from 'mdi-react/ProgressWrenchIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike } from '@sourcegraph/common'
import { IncompleteDatapointAlert } from '@sourcegraph/shared/src/schema'
import {
    Alert,
    Button,
    H4,
    Icon,
    Text,
    Popover,
    PopoverTrigger,
    PopoverContent,
    PopoverTail,
    ScrollBox,
    Link,
} from '@sourcegraph/wildcard'

import { BackendInsightSeries } from '../../../../../../core'
import { InsightInProcessError } from '../../../../../../core/backend/utils/errors'

import styles from './BackendInsightAlerts.module.scss'

interface BackendAlertOverLayProps {
    isFetchingHistoricalData: boolean
    hasNoData: boolean
    className?: string
}

export const BackendAlertOverlay: FC<BackendAlertOverLayProps> = props => {
    const { isFetchingHistoricalData, hasNoData, className } = props

    if (isFetchingHistoricalData) {
        return (
            <AlertOverlay
                title="This insight is still being processed"
                description="Datapoints shown may be undercounted."
                icon={<ProgressWrench className={classNames('mb-3')} size={33} />}
                className={className}
            />
        )
    }

    if (hasNoData) {
        return (
            <AlertOverlay
                title="No data to display"
                description="We couldn’t find any matches for this insight."
                className={className}
            />
        )
    }

    return null
}

export interface AlertOverlayProps {
    title: string
    description: string
    icon?: React.ReactNode
    className?: string
}

const AlertOverlay: React.FunctionComponent<React.PropsWithChildren<AlertOverlayProps>> = props => {
    const { title, description, icon, className } = props

    return (
        <div className={classNames(className, styles.alertOverlay)}>
            {icon && <div className={styles.alertOverlayIcon}>{icon}</div>}
            <H4 className={styles.alertOverlayTitle}>{title}</H4>
            <small className={styles.alertOverlayDescription}>{description}</small>
        </div>
    )
}

interface BackendInsightErrorAlertProps {
    error: ErrorLike
}

export const BackendInsightErrorAlert: FC<BackendInsightErrorAlertProps> = props =>
    props.error instanceof InsightInProcessError ? (
        <Alert variant="info">{props.error.message}</Alert>
    ) : (
        <ErrorAlert error={props.error} />
    )

interface InsightIncompleteAlertProps {
    alert: IncompleteDatapointAlert
}

export const InsightIncompleteAlert: FC<InsightIncompleteAlertProps> = props => {
    const { alert } = props

    return (
        <Popover>
            <PopoverTrigger as={Button} variant="icon" className={styles.alertIcon}>
                <Icon
                    aria-label="Insight is in incomplete state"
                    svgPath={mdiAlertCircleOutline}
                    color="var(--warning)"
                />
            </PopoverTrigger>

            <PopoverContent position="bottom" className={classNames(styles.alertPopover, styles.alertPopoverSmall)}>
                {getAlertMessage(alert)}{' '}
                <Link to="/help/code_insights/how-tos/Troubleshooting" target="_blank" rel="noopener">
                    Troubleshoot
                </Link>
            </PopoverContent>

            <PopoverTail size="sm" />
        </Popover>
    )
}

function getAlertMessage(alert: IncompleteDatapointAlert): string {
    switch (alert.__typename) {
        case 'TimeoutDatapointAlert':
            return 'Calculating some points on this insight exceeded the timeout limit. Results may be incomplete.'
        case 'GenericIncompleteDatapointAlert':
            return alert.reason
    }
}

interface InsightSeriesIncompleteAlertProps {
    series: BackendInsightSeries<unknown>
}
const dateFormatter = timeFormat('%B %d, %Y')

export const InsightSeriesIncompleteAlert: FC<InsightSeriesIncompleteAlertProps> = props => {
    const { series } = props

    return (
        <Popover>
            <PopoverTrigger as={Button} variant="icon" className={styles.alertIcon}>
                <Icon
                    aria-label="Insight is in incomplete state"
                    svgPath={mdiAlertCircleOutline}
                    color="var(--warning)"
                />
            </PopoverTrigger>

            <PopoverContent
                position="right"
                className={styles.alertPopover}
                focusContainerClassName={styles.alertPopoverFocusContainer}
            >
                <Text className={styles.alertDescription}>
                    Some points of this data series got errors. Results may be incomplete.{' '}
                    <Link to="/help/code_insights/how-tos/Troubleshooting" target="_blank" rel="noopener">
                        Troubleshoot
                    </Link>
                </Text>

                <ScrollBox lazyMeasurements={true} className={styles.alertPointsListScroll}>
                    <ul className={styles.alertPointsList}>
                        {series.alerts.map(alert => (
                            <li key={alert.time} className={styles.alertPoint}>
                                <div className={styles.alertPointDotContainer}>
                                    <span
                                        /* eslint-disable-next-line react/forbid-dom-props */
                                        style={{ backgroundColor: series.color }}
                                        className={styles.alertPointDot}
                                    />
                                </div>

                                <div className={styles.alertPointDescription}>
                                    {dateFormatter(new Date(alert.time))}
                                    <small>{getPointAlertMessage(alert)}</small>
                                </div>
                            </li>
                        ))}
                    </ul>
                </ScrollBox>
            </PopoverContent>

            <PopoverTail size="sm" />
        </Popover>
    )
}

function getPointAlertMessage(alert: IncompleteDatapointAlert): string {
    switch (alert.__typename) {
        case 'TimeoutDatapointAlert':
            return 'This data point exceeded the time limit.'
        case 'GenericIncompleteDatapointAlert':
            return alert.reason
    }
}
