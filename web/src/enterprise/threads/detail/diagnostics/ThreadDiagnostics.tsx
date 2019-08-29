import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { ThemeProps } from '../../../../theme'
import { useThreadDiagnostics } from './useThreadDiagnostics'
import { DiagnosticListByResource } from '../../../../diagnostics/list/byResource/DiagnosticListByResource'
import { toDiagnostic } from '../../../../../../shared/src/api/types/diagnostic'

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
export const ThreadDiagnostics: React.FunctionComponent<Props> = ({ thread, className = '', ...props }) => {
    const [diagnostics] = useThreadDiagnostics(thread)
    return (
        <div className={`campaign-diagnostics ${className}`}>
            {diagnostics === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(diagnostics) ? (
                <div className="alert alert-danger">{diagnostics.message}</div>
            ) : diagnostics.totalCount === 0 ? (
                <span className="text-muted">No diagnostics</span>
            ) : (
                <DiagnosticListByResource
                    {...props}
                    diagnostics={diagnostics.edges.map(e => ({
                        ...e.diagnostic.data,
                        ...toDiagnostic(e.diagnostic.data),
                    }))}
                    listClassName="list-group"
                />
            )}
        </div>
    )
}
