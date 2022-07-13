import { mdiAlertCircle, mdiCheckBold, mdiTimerSand, mdiTimelineClockOutline, mdiCircleOffOutline } from '@mdi/js'

import { pluralize } from '@sourcegraph/common'
import { Icon } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceStats } from '../../../../graphql-operations'

import styles from './ExecutionStatsBar.module.scss'

export const ExecutionStatsBar: React.FunctionComponent<React.PropsWithChildren<BatchSpecWorkspaceStats>> = stats => (
    <>
        <ExecutionStat>
            <Icon aria-hidden={true} className="text-danger" svgPath={mdiAlertCircle} />
            {stats.errored} {pluralize('error', stats.errored)}
        </ExecutionStat>
        <ExecutionStat>
            <Icon aria-hidden={true} className="text-success" svgPath={mdiCheckBold} />
            {stats.completed} complete
        </ExecutionStat>
        <ExecutionStat>
            <Icon aria-hidden={true} svgPath={mdiTimerSand} />
            {stats.processing} working
        </ExecutionStat>
        <ExecutionStat>
            <Icon aria-hidden={true} svgPath={mdiTimelineClockOutline} />
            {stats.queued} queued
        </ExecutionStat>
        <ExecutionStat>
            <Icon aria-hidden={true} svgPath={mdiCircleOffOutline} />
            {stats.ignored} ignored
        </ExecutionStat>
    </>
)

export const ExecutionStat: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <div className={styles.stat}>{children}</div>
)
