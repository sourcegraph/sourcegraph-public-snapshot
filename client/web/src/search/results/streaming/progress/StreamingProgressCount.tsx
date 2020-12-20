import classNames from 'classnames'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import * as React from 'react'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { StreamingProgressProps } from './StreamingProgress'

export const StreamingProgressCount: React.FunctionComponent<Pick<StreamingProgressProps, 'progress' | 'state'>> = ({
    progress,
    state,
}) => (
    <div
        className={classNames('streaming-progress__count d-flex align-items-center', {
            'streaming-progress__count--in-progress': state === 'loading',
        })}
    >
        <CalculatorIcon className="mr-2 icon-inline" />
        {progress.matchCount} {pluralize('result', progress.matchCount)} in {(progress.durationMs / 1000).toFixed(2)}s
        {progress.repositoriesCount && (
            <>
                {' '}
                from {progress.repositoriesCount} {pluralize('repository', progress.repositoriesCount, 'repositories')}
            </>
        )}
    </div>
)
