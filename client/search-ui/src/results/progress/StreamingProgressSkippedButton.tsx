import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { Button, Icon } from '@sourcegraph/wildcard'

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
                    <Button
                        className="mb-0 d-flex align-items-center text-decoration-none"
                        caret={true}
                        outline={true}
                        variant={skippedWithWarningOrError ? 'danger' : 'secondary'}
                        data-testid="streaming-progress-skipped"
                        size="sm"
                        as={DropdownToggle}
                    >
                        {skippedWithWarningOrError ? (
                            <Icon className={classNames('mr-2', styles.alertDangerIcon)} as={AlertCircleIcon} />
                        ) : null}
                        Some results excluded
                    </Button>
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
