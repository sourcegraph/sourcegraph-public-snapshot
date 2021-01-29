import classNames from 'classnames'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import * as React from 'react'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { StreamingProgressProps } from './StreamingProgress'

const abbreviateNumber = (number: number): string => {
    if (number < 1e3) {
        return number.toString()
    }
    if (number >= 1e3 && number < 1e6) {
        return (number / 1e3).toFixed(1) + 'k+'
    }
    if (number >= 1e6 && number < 1e9) {
        return (number / 1e6).toFixed(1) + 'm+'
    }
    return (number / 1e9).toFixed(1) + 'b+'
}

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
        {abbreviateNumber(progress.matchCount)} {pluralize('result', progress.matchCount)} in{' '}
        {(progress.durationMs / 1000).toFixed(2)}s
        {progress.repositoriesCount !== undefined && (
            <>
                {' '}
                from {abbreviateNumber(progress.repositoriesCount)}{' '}
                {pluralize('repository', progress.repositoriesCount, 'repositories')}
            </>
        )}
    </div>
)
