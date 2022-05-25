import React from 'react'

import { SourcegraphLogo } from '@sourcegraph/branded/src/components/SourcegraphLogo'
import { StreamingProgressCount } from '@sourcegraph/search-ui'
import { Progress, StreamingResultsState } from '@sourcegraph/shared/src/search/stream'

import styles from './Title.module.scss'

interface Props {
    state: StreamingResultsState | null
    progress: Progress
}

export const Title: React.FunctionComponent<Props> = ({ state, progress }: Props) => (
    <div className={styles.header}>
        <SourcegraphLogo className={styles.logo} />
        {state !== null && <StreamingProgressCount progress={progress} state={state} showTrace={false} />}
    </div>
)
