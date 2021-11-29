import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

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
                <ButtonDropdown isOpen={isOpen} toggle={toggleOpen}>
                    <DropdownToggle
                        className="mb-0 d-flex align-items-center text-decoration-none btn-sm"
                        caret={true}
                        color={skippedWithWarningOrError ? 'outline-danger' : 'outline-secondary'}
                        data-testid="streaming-progress-skipped"
                    >
                        {skippedWithWarningOrError ? <AlertCircleIcon className="mr-2 icon-inline" /> : null}
                        Some results excluded
                    </DropdownToggle>
                    <DropdownMenu className={styles.skippedPopover} data-testid="streaming-progress-skipped-popover">
                        <StreamingProgressSkippedPopover
                            progress={progress}
                            onSearchAgain={onSearchAgainWithPopupClose}
                        />
                    </DropdownMenu>
                </ButtonDropdown>
            )}
        </>
    )
}
