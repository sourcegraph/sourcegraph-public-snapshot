import { mdiAlertCircle, mdiCheckBold, mdiTimerSand, mdiTimelineClockOutline, mdiCircleOffOutline } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'

import { pluralize } from '@sourcegraph/common'
import { Icon, H3 } from '@sourcegraph/wildcard'

import type { BatchSpecWorkspaceStats } from '../../../../graphql-operations'

import styles from './ExecutionStatsBar.module.scss'

export const ExecutionStatsBar: React.FunctionComponent<React.PropsWithChildren<BatchSpecWorkspaceStats>> = stats => (
    <>
        <ExecutionStat>
            <Icon aria-hidden={true} className="text-danger" svgPath={mdiAlertCircle} />
            <H3 className={styles.label}>
                {stats.errored} <VisuallyHidden>workspace</VisuallyHidden> {pluralize('error', stats.errored)}
            </H3>
        </ExecutionStat>
        <ExecutionStat>
            <Icon aria-hidden={true} className="text-success" svgPath={mdiCheckBold} />
            <H3 className={styles.label}>
                {stats.completed} <VisuallyHidden>{pluralize('workspace', stats.completed)}</VisuallyHidden> complete
            </H3>
        </ExecutionStat>
        <ExecutionStat>
            <Icon aria-hidden={true} svgPath={mdiTimerSand} />
            <H3 className={styles.label}>
                {stats.processing} <VisuallyHidden>{pluralize('workspace', stats.processing)}</VisuallyHidden> working
            </H3>
        </ExecutionStat>
        <ExecutionStat>
            <Icon aria-hidden={true} svgPath={mdiTimelineClockOutline} />
            <H3 className={styles.label}>
                {stats.queued} <VisuallyHidden>{pluralize('workspace', stats.queued)}</VisuallyHidden> queued
            </H3>
        </ExecutionStat>
        <ExecutionStat>
            <Icon aria-hidden={true} svgPath={mdiCircleOffOutline} />
            <H3 className={styles.label}>
                {stats.ignored} <VisuallyHidden>{pluralize('workspace', stats.ignored)}</VisuallyHidden> ignored
            </H3>
        </ExecutionStat>
    </>
)

export const ExecutionStat: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <div className={styles.stat}>{children}</div>
)
