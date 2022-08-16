import React, { useCallback, useMemo, useState } from 'react'

import { mdiAlertCircle, mdiChevronDown, mdiInformationOutline } from '@mdi/js'

import { useCoreWorkflowImprovementsEnabled } from '@sourcegraph/shared/src/settings/useCoreWorkflowImprovementsEnabled'
import { Button, Popover, PopoverContent, PopoverTrigger, Position, Icon } from '@sourcegraph/wildcard'

import { StreamingProgressProps } from './StreamingProgress'
import { CountContent, getProgressText } from './StreamingProgressCount'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

import styles from './StreamingProgressSkippedButton.module.scss'

export const StreamingProgressSkippedButton: React.FunctionComponent<
    React.PropsWithChildren<Pick<StreamingProgressProps, 'progress' | 'onSearchAgain'>>
> = ({ progress, onSearchAgain }) => {
    const [coreWorkflowImprovementsEnabled] = useCoreWorkflowImprovementsEnabled()

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

    const progressText = getProgressText(progress)

    return (
        <>
            {(coreWorkflowImprovementsEnabled || progress.skipped.length > 0) && (
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
                            <Icon aria-hidden={true} className="mr-2" svgPath={mdiAlertCircle} />
                        ) : coreWorkflowImprovementsEnabled ? (
                            <Icon aria-hidden={true} className="mr-2" svgPath={mdiInformationOutline} />
                        ) : null}
                        {coreWorkflowImprovementsEnabled ? (
                            <CountContent progressText={progressText} />
                        ) : (
                            <>Some results excluded </>
                        )}
                        <Icon aria-hidden={true} data-caret={true} className="mr-0" svgPath={mdiChevronDown} />
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
