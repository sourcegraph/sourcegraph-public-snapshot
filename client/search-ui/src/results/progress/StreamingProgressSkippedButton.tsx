import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { Button, Popover, PopoverContent, PopoverTrigger, Position } from '@sourcegraph/wildcard'

import { StreamingProgressProps } from './StreamingProgress'
import styles from './StreamingProgressSkippedButton.module.scss'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

export const StreamingProgressSkippedButton: React.FunctionComponent<
    Pick<StreamingProgressProps, 'progress' | 'onSearchAgain'>
> = ({ progress, onSearchAgain }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

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
                <Popover isOpen={isOpen} onOpenChange={toggleOpen}>
                    <PopoverTrigger
                        className={classNames('mb-0 d-flex align-items-center text-decoration-none', styles.skippedBtn)}
                        size="sm"
                        variant={skippedWithWarningOrError ? 'danger' : 'secondary'}
                        outline={true}
                        data-testid="streaming-progress-skipped"
                        as={Button}
                        aria-expanded={isOpen}
                    >
                        {skippedWithWarningOrError ? (
                            <AlertCircleIcon className={classNames('mr-2 icon-inline', styles.alertDangerIcon)} />
                        ) : null}
                        Some results excluded
                        {isOpen ? (
                            <MenuUpIcon className="icon-inline caret" />
                        ) : (
                            <MenuDownIcon className="icon-inline caret" />
                        )}
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
