import React from 'react'

import { StreamingProgressCount } from '@sourcegraph/search-ui'
import { Progress, StreamingResultsState } from '@sourcegraph/shared/src/search/stream'

import styles from './StatusBar.module.scss'

interface Props {
    progress: Progress
    progressState: StreamingResultsState | null
    authState: 'initial' | 'validating' | 'success' | 'failure'
}

export const StatusBar: React.FunctionComponent<Props> = ({ progress, progressState, authState }: Props) => (
    <div className={styles.statusBar}>
        {progressState !== null && (
            <StreamingProgressCount
                progress={progress}
                state={progressState}
                showTrace={false}
                className={styles.progressCount}
            />
        )}
        <div className={styles.spacer} />
        <div>Auth state: {authState}</div>
    </div>
)
