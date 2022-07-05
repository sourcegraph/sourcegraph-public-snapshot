import * as React from 'react'

import { mdiClipboardPulseOutline, mdiInformationOutline } from '@mdi/js'
import VisuallyHidden from '@reach/visually-hidden'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { Progress } from '@sourcegraph/shared/src/search/stream'
import { Link, Icon, Tooltip } from '@sourcegraph/wildcard'

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
> = ({ progress, state, showTrace, className = '' }) => {
    const isLoading = state === 'loading'
    const contentWithoutTimeUnit =
        `${abbreviateNumber(progress.matchCount)}` +
        `${limitHit(progress) ? '+' : ''} ${pluralize('result', progress.matchCount)} in ` +
        `${(progress.durationMs / 1000).toFixed(2)}`
    const content = `${contentWithoutTimeUnit}s`
    const readingContent = `${contentWithoutTimeUnit} seconds`

    return (
        <>
            {isLoading && <VisuallyHidden aria-live="polite">Searching</VisuallyHidden>}
            <small
                className={classNames(
                    'd-flex align-items-center',
                    className,
                    styles.count,
                    isLoading && styles.countInProgress
                )}
                data-testid="streaming-progress-count"
            >
                {/*
                    Span wrapper needed to avoid VisuallyHidden creating a scrollable overflow in Chrome.
                    Related bug: https://bugs.chromium.org/p/chromium/issues/detail?id=1154640#c15
                 */}
                <span className="position-relative">
                    <VisuallyHidden aria-live="polite">{readingContent}</VisuallyHidden>
                </span>
                <span aria-hidden={true}>{content}</span>
                {progress.repositoriesCount !== undefined && (
                    <Tooltip
                        content={`From ${abbreviateNumber(progress.repositoriesCount)} ${pluralize(
                            'repository',
                            progress.repositoriesCount,
                            'repositories'
                        )}`}
                    >
                        <Icon
                            className="ml-1"
                            svgPath={mdiInformationOutline}
                            tabIndex={0}
                            aria-label={`From ${abbreviateNumber(progress.repositoriesCount)} ${pluralize(
                                'repository',
                                progress.repositoriesCount,
                                'repositories'
                            )}`}
                        />
                    </Tooltip>
                )}
            </small>
            {showTrace && progress.trace && (
                <small className={classNames('d-flex align-items-center', className, styles.count)}>
                    <Link to={progress.trace}>
                        <Icon aria-hidden={true} className="mr-2" svgPath={mdiClipboardPulseOutline} />
                        View trace
                    </Link>
                </small>
            )}
        </>
    )
}
