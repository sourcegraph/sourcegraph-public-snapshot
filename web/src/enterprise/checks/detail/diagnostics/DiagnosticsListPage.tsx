import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { withQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThemeProps } from '../../../../theme'
import { DiagnosticsList } from '../../../tasks/list/DiagnosticsList'
import { useDiagnostics } from './detail/useDiagnostics'
import { DiagnosticQueryBuilder } from './DiagnosticQueryBuilder'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    baseDiagnosticQuery: sourcegraph.DiagnosticQuery

    className?: string
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * A page that lists diagnostics.
 */
export const DiagnosticsListPage = withQueryParameter<Props>(
    ({ baseDiagnosticQuery, query, onQueryChange, className = '', extensionsController, ...props }) => {
        const parsedQuery = parseDiagnosticQuery(query, baseDiagnosticQuery)
        // tslint:disable-next-line: react-hooks-nesting
        const diagnosticsOrError = useDiagnostics(extensionsController, parsedQuery)
        return (
            <div className={`check-diagnostics-page ${className}`}>
                {isErrorLike(diagnosticsOrError) ? (
                    <div className="container">
                        <div className="alert alert-danger mt-2">{diagnosticsOrError.message}</div>
                    </div>
                ) : diagnosticsOrError === LOADING ? (
                    <div className="container">
                        <LoadingSpinner className="mt-3" />
                    </div>
                ) : diagnosticsOrError.length === 0 ? (
                    <div className="container">
                        <p className="p-2 mb-0 text-muted">No diagnostics found.</p>
                    </div>
                ) : (
                    <>
                        <DiagnosticQueryBuilder
                            parsedQuery={parsedQuery}
                            query={query}
                            onQueryChange={onQueryChange}
                            className="container my-3"
                        />
                        <DiagnosticsList
                            {...props}
                            diagnosticsOrError={diagnosticsOrError}
                            itemClassName="container-fluid"
                            extensionsController={extensionsController}
                        />
                    </>
                )}
            </div>
        )
    }
)

function parseDiagnosticQuery(query: string, base?: sourcegraph.DiagnosticQuery): sourcegraph.DiagnosticQuery {
    return base || { type: 'TODO!(sqs)' } // TODO!(sqs)
}
