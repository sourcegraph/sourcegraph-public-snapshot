import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import * as React from 'react'
import { Button, Popover, PopoverBody } from 'reactstrap'
import { defaultProgress, StreamingProgressProps } from './StreamingProgress'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

export const StreamingProgressSkippedButton: React.FunctionComponent<StreamingProgressProps> = ({
    progress = defaultProgress,
}) => {
    const [isOpen, setIsOpen] = React.useState(false)
    const toggleOpen = React.useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

    const skippedWithWarning = progress.skipped.some(skipped => skipped.severity === 'warn')

    return (
        <>
            {progress.skipped.length > 0 && (
                <>
                    <Button
                        className={classNames('streaming-progress__skipped p-2 mb-0 d-flex align-items-center', {
                            'alert alert-danger': skippedWithWarning,
                        })}
                        color={skippedWithWarning ? 'danger' : 'secondary'}
                        onClick={toggleOpen}
                        id="streaming-progress__skipped"
                    >
                        {skippedWithWarning ? (
                            <AlertCircleIcon className="mr-2" />
                        ) : (
                            <InformationOutlineIcon className="mr-2" />
                        )}
                        Some repositories excluded
                        <MenuDownIcon className="icon-inline" />
                    </Button>
                    <Popover
                        placement="bottom-start"
                        isOpen={isOpen}
                        toggle={toggleOpen}
                        target="streaming-progress__skipped"
                        hideArrow={true}
                    >
                        <PopoverBody>
                            <StreamingProgressSkippedPopover progress={progress} />
                        </PopoverBody>
                    </Popover>
                </>
            )}
        </>
    )
}
