import type { FC, ReactNode } from 'react'

import { mdiAlertCircle as mdiAlertCircleOutline } from '@mdi/js'
import classNames from 'classnames'
import { timeFormat } from 'd3-time-format'
import ProgressWrench from 'mdi-react/ProgressWrenchIcon'

import type { ErrorLike } from '@sourcegraph/common'
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
    ErrorAlert,
} from '@sourcegraph/wildcard'

import type { BackendInsightSeries } from '../../../../../../core'
import { InsightInProcessError } from '../../../../../../core/backend/utils/errors'
import type { IncompleteDatapointAlert } from '../../../../../../core/types/insight/common'

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
                icon={<ProgressWrench size={33} />}
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
    description?: ReactNode
    icon?: ReactNode
    className?: string
}

const AlertOverlay: FC<AlertOverlayProps> = props => {
    const { title, description, icon, className } = props

    return (
        <div className={classNames(className, styles.alertOverlay)}>
            {icon && <div className={styles.alertOverlayIcon}>{icon}</div>}
            <H4 className={styles.alertOverlayTitle}>{title}</H4>
            {description && <small className={styles.alertOverlayDescription}>{description}</small>}
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
                    color="var(--icon-color)"
                />
            </PopoverTrigger>

            <PopoverContent position="bottom" className={classNames(styles.alertPopover, styles.alertPopoverSmall)}>
                {getAlertMessage(alert)}{' '}
                <Link to="/help/code_insights/references/incomplete_data_points" target="_blank" rel="noopener">
                    Troubleshoot
                </Link>
            </PopoverContent>

            <PopoverTail size="sm" />
        </Popover>
    )
}

function getAlertMessage(alert: IncompleteDatapointAlert): ReactNode {
    switch (alert.__typename) {
        case 'TimeoutDatapointAlert': {
            return (
                <>
                    Some points of this data series <b>exceeded the timeout limit</b>. Results may be incomplete.
                </>
            )
        }
        case 'GenericIncompleteDatapointAlert': {
            // Since BE doesn't handle insight level alerts properly we can't use
            // alert.reason message here but hardcoded on the client error message.
            return 'Some points of this data series encountered an error. Results may be incomplete.'
        }
    }
}

interface InsightSeriesIncompleteAlertProps {
    series: BackendInsightSeries<unknown>
    className?: string
}

const dateFormatter = timeFormat('%B %d, %Y')

export const InsightSeriesIncompleteAlert: FC<InsightSeriesIncompleteAlertProps> = props => {
    const { series, className } = props

    const timeoutAlerts = series.alerts.filter(alert => alert.__typename === 'TimeoutDatapointAlert')
    const otherAlerts = series.alerts.filter(alert => alert.__typename !== 'TimeoutDatapointAlert')

    return (
        <Popover>
            <PopoverTrigger as={Button} variant="icon" className={classNames(className, styles.alertIcon)}>
                <Icon
                    aria-label="Insight is in incomplete state"
                    svgPath={mdiAlertCircleOutline}
                    color="var(--icon-color)"
                />
            </PopoverTrigger>

            <PopoverContent
                position="right"
                className={styles.alertPopover}
                focusContainerClassName={styles.alertPopoverFocusContainer}
            >
                <Text className={styles.alertDescription}>
                    Results for some points of this data series may be incomplete.{' '}
                    <Link to="/help/code_insights/references/incomplete_data_points" target="_blank" rel="noopener">
                        Troubleshoot
                    </Link>
                </Text>

                <ScrollBox lazyMeasurements={true} className={styles.alertPointsListScroll}>
                    {timeoutAlerts.length > 0 && (
                        <>
                            <Text className={styles.alertPointSectionTitle}>Exceeded the timeout limit:</Text>
                            <ul className={styles.alertPointsList}>
                                {timeoutAlerts.map(alert => (
                                    <li key={alert.time} className={styles.alertPoint}>
                                        <div className={styles.alertPointDotContainer}>
                                            <span
                                                /* eslint-disable-next-line react/forbid-dom-props */
                                                style={{ backgroundColor: series.color }}
                                                className={styles.alertPointDot}
                                            />
                                        </div>

                                        {dateFormatter(new Date(alert.time))}
                                    </li>
                                ))}
                            </ul>
                        </>
                    )}

                    {otherAlerts.length > 0 && (
                        <>
                            <Text className={styles.alertPointSectionTitle}>Unable to calculate:</Text>
                            <ul className={styles.alertPointsList}>
                                {otherAlerts.map(alert => (
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
                        </>
                    )}
                </ScrollBox>
            </PopoverContent>

            <PopoverTail size="sm" />
        </Popover>
    )
}

function getPointAlertMessage(alert: IncompleteDatapointAlert): string {
    switch (alert.__typename) {
        case 'TimeoutDatapointAlert': {
            return 'This data point exceeded the time limit.'
        }
        case 'GenericIncompleteDatapointAlert': {
            return alert.reason
        }
    }
}
