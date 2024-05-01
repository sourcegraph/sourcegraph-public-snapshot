import { useCallback, useMemo, useState, type FC } from 'react'

import { mdiAlertCircle, mdiChevronDown, mdiInformationOutline } from '@mdi/js'

import type { Progress } from '@sourcegraph/shared/src/search/stream'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Popover, PopoverContent, PopoverTrigger, Position, Icon } from '@sourcegraph/wildcard'

import { CountContent } from './StreamingProgressCount'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'
import { getProgressText } from './utils'

import styles from './StreamingProgressSkippedButton.module.scss'

interface StreamingProgressSkippedButtonProps extends TelemetryProps, TelemetryV2Props {
    query: string
    progress: Progress
    isSearchJobsEnabled?: boolean
    onSearchAgain: (additionalFilters: string[]) => void
}

export const StreamingProgressSkippedButton: FC<StreamingProgressSkippedButtonProps> = props => {
    const { query, progress, isSearchJobsEnabled, onSearchAgain, telemetryService, telemetryRecorder } = props
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
                ) : (
                    <Icon aria-hidden={true} className="mr-2" svgPath={mdiInformationOutline} />
                )}
                <CountContent progressText={progressText} />
                <Icon aria-hidden={true} data-caret={true} className="mr-0" svgPath={mdiChevronDown} />
            </PopoverTrigger>
            <PopoverContent
                position={Position.bottomStart}
                className={styles.skippedPopover}
                data-testid="streaming-progress-skipped-popover"
            >
                <StreamingProgressSkippedPopover
                    query={query}
                    progress={progress}
                    isSearchJobsEnabled={isSearchJobsEnabled}
                    onSearchAgain={onSearchAgainWithPopupClose}
                    telemetryService={telemetryService}
                    telemetryRecorder={telemetryRecorder}
                />
            </PopoverContent>
        </Popover>
    )
}
