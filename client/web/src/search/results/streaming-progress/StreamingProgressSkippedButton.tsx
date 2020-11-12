import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import React, { useCallback, useState, useMemo, useEffect } from 'react'
import { Button, Popover, PopoverBody } from 'reactstrap'
import { defaultProgress, StreamingProgressProps } from './StreamingProgress'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

export const StreamingProgressSkippedButton: React.FunctionComponent<StreamingProgressProps> = ({
    progress = defaultProgress,
    popoverOpen = false,
}) => {
    const [isOpen, setIsOpen] = useState(popoverOpen)
    const toggleOpen = useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

    // Shouldn't render popover initially, wait until component is initialized to render it.
    const [renderPopover, setRenderPopover] = useState(false)
    useEffect(() => {
        setRenderPopover(true)
    }, [])

    const skippedWithWarning = useMemo(() => progress.skipped.some(skipped => skipped.severity === 'warn'), [progress])

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
                    {renderPopover && (
                        <Popover
                            placement="bottom-start"
                            isOpen={isOpen}
                            toggle={toggleOpen}
                            target="streaming-progress__skipped"
                            hideArrow={true}
                        >
                            <PopoverBody className="streaming-progress__skipped__popover">
                                <StreamingProgressSkippedPopover progress={progress} />
                            </PopoverBody>
                        </Popover>
                    )}
                </>
            )}
        </>
    )
}
