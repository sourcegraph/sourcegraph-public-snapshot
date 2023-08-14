import * as React from 'react'

import { mdiInformationOutline } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import type { Progress } from '@sourcegraph/shared/src/search/stream'
import { Icon, Tooltip } from '@sourcegraph/wildcard'

import type { StreamingProgressProps } from './StreamingProgress'
import { limitHit } from './utils'

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

export const getProgressText = (progress: Progress): { visibleText: string; readText: string } => {
    const contentWithoutTimeUnit =
        `${abbreviateNumber(progress.matchCount)}` +
        `${limitHit(progress) ? '+' : ''} ${pluralize('result', progress.matchCount)} in ` +
        `${(progress.durationMs / 1000).toFixed(2)}`
    const visibleText = `${contentWithoutTimeUnit}s`
    const readText = `${contentWithoutTimeUnit} seconds`
    return { visibleText, readText }
}

export const StreamingProgressCount: React.FunctionComponent<
    React.PropsWithChildren<
        Pick<StreamingProgressProps, 'progress' | 'state'> & { className?: string; hideIcon?: boolean }
    >
> = ({ progress, state, className = '', hideIcon = false }) => {
    const isLoading = state === 'loading'
    const progressText = getProgressText(progress)

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
                <CountContent progressText={progressText} />
                {!hideIcon && progress.repositoriesCount !== undefined && (
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
        </>
    )
}

export const CountContent: React.FunctionComponent<{ progressText: { visibleText: string; readText: string } }> = ({
    progressText,
}) => (
    <>
        {/*
        Span wrapper needed to avoid VisuallyHidden creating a scrollable overflow in Chrome.
        Related bug: https://bugs.chromium.org/p/chromium/issues/detail?id=1154640#c15
        */}
        <span className="position-relative">
            <VisuallyHidden aria-live="polite">{progressText.readText}</VisuallyHidden>
        </span>
        <span aria-hidden={true}>{progressText.visibleText}</span>
    </>
)
