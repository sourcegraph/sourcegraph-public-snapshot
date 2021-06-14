import classNames from 'classnames'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import ClipboardPulseOutlineIcon from 'mdi-react/ClipboardPulseOutlineIcon'
import * as React from 'react'

import { Progress } from '@sourcegraph/shared/src/search/stream'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { StreamingProgressProps } from './StreamingProgress'

const abbreviateNumber = (number: number): string => {
    if (number < 1e3) {
        return number.toString()
    }
    if (number >= 1e3 && number < 1e6) {
        return (number / 1e3).toFixed(1) + 'k'
    }
    if (number >= 1e6 && number < 1e9) {
        return (number / 1e6).toFixed(1) + 'm'
    }
    return (number / 1e9).toFixed(1) + 'b'
}

const limitHit = (progress: Progress): boolean => progress.skipped.some(skipped => skipped.reason.indexOf('-limit') > 0)

export const StreamingProgressCount: React.FunctionComponent<
    Pick<StreamingProgressProps, 'progress' | 'state' | 'showTrace'> & { className?: string }
> = ({ progress, state, showTrace, className = '' }) => (
    <>
        <small
            className={classNames(className, 'streaming-progress__count d-flex align-items-center', {
                'streaming-progress__count--in-progress': state === 'loading',
            })}
        >
            <CalculatorIcon className="mr-2 icon-inline streaming-progress__count-icon" />
            {abbreviateNumber(progress.matchCount)}
            {limitHit(progress) ? '+' : ''} {pluralize('result', progress.matchCount)} in{' '}
            {(progress.durationMs / 1000).toFixed(2)}s
            {progress.repositoriesCount !== undefined && (
                <>
                    {' '}
                    from {abbreviateNumber(progress.repositoriesCount)}{' '}
                    {pluralize('repository', progress.repositoriesCount, 'repositories')}
                </>
            )}
        </small>
        {showTrace && progress.trace && (
            <small className="d-flex ml-2">
                <a href={progress.trace}>
                    <ClipboardPulseOutlineIcon className="mr-2 icon-inline" />
                    View trace
                </a>
            </small>
        )}
    </>
)
