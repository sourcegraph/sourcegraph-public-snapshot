import * as React from 'react'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import { pluralize } from '../../../../../shared/src/util/strings'
import { Progress } from '../../stream'
import { Button } from 'reactstrap'

interface Props {
    progress: Progress
}

const defaultProgress: Progress = {
    done: true,
    durationMs: 0,
    matchCount: 0,
    skipped: [],
}

export const StreamingProgress: React.FunctionComponent<Props> = ({ progress = defaultProgress }) => {
    const [isOpen, setIsOpen] = React.useState(false)
    const toggleOpen = React.useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

    const skippedWithWarning = progress.skipped.some(skipped => skipped.severity === 'warn')

    return (
        <div className="d-flex streaming-progress">
            <div
                className={classNames('streaming-progress__count p-2 d-flex align-items-center', {
                    'streaming-progress__count--in-progress': !progress.done,
                })}
            >
                <CalculatorIcon className="mr-2" />
                {progress.matchCount} {pluralize('result', progress.matchCount)} in{' '}
                {(progress.durationMs / 1000).toFixed(2)}s
                {progress.repositoriesCount && (
                    <>
                        {' '}
                        from {progress.repositoriesCount}{' '}
                        {pluralize('repository', progress.repositoriesCount, 'repositories')}
                    </>
                )}
            </div>

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
        </div>
    )
}
