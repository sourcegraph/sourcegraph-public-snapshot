import * as React from 'react'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import { Progress } from '../../stream'
import { pluralize } from '../../../../../shared/src/util/strings'
import classNames from 'classnames'

interface Props {
    progress: Progress
}

const defaultProgress: Progress = {
    done: false,
    durationMs: 0,
    matchCount: 0,
    skipped: [],
}

export const StreamingProgress: React.FunctionComponent<Props> = ({ progress = defaultProgress }) => (
    <div className="d-flex">
        <div
            className={classNames('streaming-progress p-2 d-flex align-items-center', {
                'streaming-progress--in-progress': !progress.done,
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
    </div>
)
