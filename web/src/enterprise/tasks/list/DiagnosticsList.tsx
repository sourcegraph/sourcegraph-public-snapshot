import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { DiagnosticInfo } from '../../threadsOLD/detail/backend'
import { TasksAreaContext } from '../global/TasksArea'
import { DiagnosticsListItem } from './item/DiagnosticsListItem'

export interface TasksListContext {
    itemClassName?: string
}

const LOADING: 'loading' = 'loading'

interface Props
    extends Partial<QueryParameterProps>,
        TasksListContext,
        TasksAreaContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    diagnosticsOrError: typeof LOADING | DiagnosticInfo[] | ErrorLike

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The list of diagnostics with a header.
 */
export const DiagnosticsList: React.FunctionComponent<Props> = ({
    diagnosticsOrError,
    className = '',
    itemClassName,
    extensionsController,
    ...props
}) => (
    <div className={`tasks-list ${className}`}>
        {isErrorLike(diagnosticsOrError) ? (
            <div className={itemClassName}>
                <div className="alert alert-danger mt-2">{diagnosticsOrError.message}</div>
            </div>
        ) : diagnosticsOrError === LOADING ? (
            <div className={itemClassName}>
                <LoadingSpinner className="mt-3" />
            </div>
        ) : diagnosticsOrError.length === 0 ? (
            <div className={itemClassName}>
                <p className="p-2 mb-0 text-muted">No diagnostics found.</p>
            </div>
        ) : (
            <ul className="list-group list-group-flush mb-0">
                {diagnosticsOrError.map((task, i) => (
                    <li key={i} className={`list-group-item px-0 ${i === 0 ? 'border-top-0' : ''}`}>
                        <DiagnosticsListItem
                            {...props}
                            key={JSON.stringify(task)}
                            diagnostic={task}
                            className={itemClassName}
                            extensionsController={extensionsController}
                        />
                    </li>
                ))}
            </ul>
        )}
        <style>
            {/* HACK TODO!(sqs) */}
            {'.tasks-list .markdown pre,.tasks-list .markdown code {margin:0;padding:0;background-color:transparent;}'}
        </style>
    </div>
)
