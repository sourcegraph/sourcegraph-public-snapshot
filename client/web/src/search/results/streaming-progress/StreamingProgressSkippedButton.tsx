import * as React from 'react'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import { Button } from 'reactstrap'
import { StreamingProgressProps } from './StreamingProgress'

export const StreamingProgressSkippedButton: React.FunctionComponent<StreamingProgressProps> = ({ progress }) => {
    const [isOpen, setIsOpen] = React.useState(false)
    const toggleOpen = React.useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

    const skippedWithWarning = progress.skipped.some(skipped => skipped.severity === 'warn')

    return (
        <>
            {progress.skipped.length > 0 && (
                <Button
                    className={classNames('streaming-progress__skipped p-2 mb-0 d-flex align-items-center', {
                        'alert alert-danger': skippedWithWarning,
                    })}
                    color={skippedWithWarning ? 'danger' : 'secondary'}
                    onClick={toggleOpen}
                >
                    {skippedWithWarning ? (
                        <AlertCircleIcon className="mr-2" />
                    ) : (
                        <InformationOutlineIcon className="mr-2" />
                    )}
                    Some repositories excluded
                    <MenuDownIcon className="icon-inline" />
                </Button>
            )}
        </>
    )
}
