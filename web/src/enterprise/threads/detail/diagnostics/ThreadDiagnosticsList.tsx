import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { ThemeProps } from '../../../../theme'
import { ThreadDiagnosticListItem } from './ThreadDiagnosticListItem'
import { useThreadDiagnostics } from './useThreadDiagnostics'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    thread: Pick<GQL.IThread, 'id'>

    className?: string
    history: H.History
    location: H.Location
}

const LOADING = 'loading' as const

/**
 * A list of diagnostics in a thread.
 */
export const ThreadDiagnosticsList: React.FunctionComponent<Props> = ({ thread, className = '', ...props }) => {
    const [diagnostics] = useThreadDiagnostics(thread)
    return (
        <div className={`thread-diagnostics-list ${className}`}>
            <ul className="list-group mb-4">
                {diagnostics === LOADING ? (
                    <LoadingSpinner className="icon-inline mt-3" />
                ) : isErrorLike(diagnostics) ? (
                    <div className="alert alert-danger mt-3">{diagnostics.message}</div>
                ) : diagnostics.totalCount === 0 ? (
                    <span className="text-muted">No diagnostics.</span>
                ) : (
                    diagnostics.edges.map((edge, i) => (
                        <li key={edge.id} className="list-group-item p-0">
                            <ThreadDiagnosticListItem {...props} threadDiagnostic={edge} />
                        </li>
                    ))
                )}
            </ul>
        </div>
    )
}
