import React from 'react'

import { StreamingProgressCount } from '@sourcegraph/branded'
import type { Progress, StreamingResultsState } from '@sourcegraph/shared/src/search/stream'

import styles from './StatusBar.module.scss'

interface Props {
    progress: Progress
    progressState: StreamingResultsState | null
    authState: 'initial' | 'validating' | 'success' | 'failure'
}

export const StatusBar: React.FunctionComponent<Props> = props => (
    <div className={styles.statusBar}>
        {props.progressState !== null &&
            (props.progressState === 'error' ? (
                ''
            ) : (
                <StreamingProgressCount
                    progress={props.progress}
                    state={props.progressState}
                    className={styles.progressCount}
                />
            ))}
        <div className={styles.spacer} />
    </div>
)
