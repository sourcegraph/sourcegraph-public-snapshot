import * as React from 'react'

import classNames from 'classnames'
import ClipboardPulseOutlineIcon from 'mdi-react/ClipboardPulseOutlineIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'

import { pluralize } from '@sourcegraph/common'
import { Progress } from '@sourcegraph/shared/src/search/stream'
import { Link, Icon } from '@sourcegraph/wildcard'

import { StreamingProgressProps } from './StreamingProgress'

import styles from './StreamingProgressCount.module.scss'

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
    React.PropsWithChildren<Pick<StreamingProgressProps, 'progress' | 'state' | 'showTrace'> & { className?: string }>
> = ({ progress, state, showTrace, className = '' }) => (
    <>
        <small
            className={classNames(
                'd-flex align-items-center',
                className,
                styles.count,
                state === 'loading' && styles.countInProgress
            )}
            data-testid="streaming-progress-count"
        >
            {abbreviateNumber(progress.matchCount)}
            {limitHit(progress) ? '+' : ''} {pluralize('result', progress.matchCount)} in{' '}
            {(progress.durationMs / 1000).toFixed(2)}s
            {progress.repositoriesCount !== undefined && (
                <Icon
                    role="img"
                    className="ml-1"
                    data-tooltip={`From ${abbreviateNumber(progress.repositoriesCount)} ${pluralize(
                        'repository',
                        progress.repositoriesCount,
                        'repositories'
                    )}`}
                    as={InformationOutlineIcon}
                    tabIndex={0}
                    aria-label={`From ${abbreviateNumber(progress.repositoriesCount)} ${pluralize(
                        'repository',
                        progress.repositoriesCount,
                        'repositories'
                    )}`}
                />
            )}
        </small>
        {showTrace && progress.trace && (
            <small className="d-flex ml-2">
                <Link to={progress.trace}>
                    <Icon role="img" aria-hidden={true} className="mr-2" as={ClipboardPulseOutlineIcon} />
                    View trace
                </Link>
            </small>
        )}
    </>
)
