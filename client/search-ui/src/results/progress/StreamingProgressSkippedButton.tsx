import React, { useCallback, useMemo, useState } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'

import { Button, Popover, PopoverContent, PopoverTrigger, Position, Icon } from '@sourcegraph/wildcard'

import { StreamingProgressProps } from './StreamingProgress'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

import styles from './StreamingProgressSkippedButton.module.scss'

export const StreamingProgressSkippedButton: React.FunctionComponent<
    React.PropsWithChildren<Pick<StreamingProgressProps, 'progress' | 'onSearchAgain'>>
> = ({ progress, onSearchAgain }) => {
    const [isOpen, setIsOpen] = useState(false)

    const skippedWithWarningOrError = useMemo(
        () => progress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error'),
        [progress]
    )

    const onSearchAgainWithPopupClose = useCallback(
        (filters: string[]) => {
            setIsOpen(false)
            onSearchAgain(filters)
        },
        [setIsOpen, onSearchAgain]
    )

    return (
        <>
            {progress.skipped.length > 0 && (
                <Popover isOpen={isOpen} onOpenChange={event => setIsOpen(event.isOpen)}>
                    <PopoverTrigger
                        className="mb-0 d-flex align-items-center text-decoration-none"
                        size="sm"
                        variant={skippedWithWarningOrError ? 'danger' : 'secondary'}
                        outline={true}
                        data-testid="streaming-progress-skipped"
                        as={Button}
                        aria-expanded={isOpen}
                        aria-label="Open excluded results"
                    >
                        {skippedWithWarningOrError ? (
                            <Icon role="img" aria-hidden={true} className="mr-2" as={AlertCircleIcon} />
                        ) : null}
                        Some results excluded{' '}
                        <Icon role="img" aria-hidden={true} data-caret={true} className="mr-0" as={ChevronDownIcon} />
                    </PopoverTrigger>
                    <PopoverContent
                        position={Position.bottomStart}
                        className={styles.skippedPopover}
                        data-testid="streaming-progress-skipped-popover"
                    >
                        <StreamingProgressSkippedPopover
                            progress={progress}
                            onSearchAgain={onSearchAgainWithPopupClose}
                        />
                    </PopoverContent>
                </Popover>
            )}
        </>
    )
}
