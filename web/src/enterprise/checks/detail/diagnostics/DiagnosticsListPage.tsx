import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { DiagnosticWithType } from '../../../../../../shared/src/api/client/services/diagnosticService'
import { Action } from '../../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { withQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThemeProps } from '../../../../theme'
import { ChangesetPlanOperation } from '../../../changesets/plan/plan'
import { DiagnosticsListItem } from '../../../tasks/list/item/DiagnosticsListItem'
import {
    diagnosticID,
    DiagnosticInfo,
    diagnosticQueryForSingleDiagnostic,
    diagnosticQueryKey,
    diagnosticQueryMatcher,
} from '../../../threads/detail/backend'
import { CheckAreaContext } from '../CheckArea'
import { useDiagnostics } from './detail/useDiagnostics'
import { parseDiagnosticQuery } from './diagnosticQuery'
import { DiagnosticQueryBuilder } from './DiagnosticQueryBuilder'
import { DiagnosticsBatchActions } from './DiagnosticsBatchActions'
import { ChangesetPlanProps } from './useChangesetPlan'

interface Props
    extends Pick<ChangesetPlanProps, 'changesetPlan' | 'onChangesetPlanBatchActionClick'>,
        Pick<CheckAreaContext, 'checkProvider'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    getSelectedActionForDiagnostic: (diagnostic: DiagnosticWithType) => ChangesetPlanOperation | null
    onActionSelect: (diagnostic: DiagnosticInfo, action: Action | null) => void

    defaultQuery: string

    className?: string
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * A page that lists diagnostics.
 */
export const DiagnosticsListPage = withQueryParameter<Props>(
    ({
        getSelectedActionForDiagnostic,
        onActionSelect,
        defaultQuery,
        changesetPlan,
        onChangesetPlanBatchActionClick,
        checkProvider,
        query,
        onQueryChange,
        className = '',
        extensionsController,
        ...props
    }) => {
        const parsedQuery = parseDiagnosticQuery(query)
        // tslint:disable-next-line: react-hooks-nesting
        const diagnosticsOrError = useDiagnostics(extensionsController, parsedQuery.type)

        const filteredDiagnosticsOrError =
            diagnosticsOrError !== LOADING && !isErrorLike(diagnosticsOrError)
                ? diagnosticsOrError.filter(diagnosticQueryMatcher(parsedQuery))
                : []

        return (
            <div className={`diagnostics-list-page ${className}`}>
                <style>{`.diagnostics-list-page code, .diagnostics-list-page table, .diagnostics-list-page pre, .diagnostics-list-page .markdown, .diagnostics-list-page aside { line-height: 17.25px; background-color: black !important; margin: 0; padding: 0; border-spacing: 0; } .diagnostics-list-page pre { margin-top: 1px !important; margin-left: 9px !important;}`}</style>
                {isErrorLike(diagnosticsOrError) ? (
                    <div className="container">
                        <div className="alert alert-danger mt-2">{diagnosticsOrError.message}</div>
                    </div>
                ) : diagnosticsOrError === LOADING ? (
                    <div className="container">
                        <LoadingSpinner className="mt-3" />
                    </div>
                ) : (
                    <>
                        <div className="diagnostics-list-page__toolbar bg-body border-bottom p-3">
                            <DiagnosticQueryBuilder
                                {...props}
                                defaultQuery={defaultQuery}
                                diagnostics={diagnosticsOrError}
                                changesetPlan={changesetPlan}
                                query={query}
                                onQueryChange={onQueryChange}
                            />
                            <div className="d-flex align-items-center mt-3 ml-3">
                                <DiagnosticsBatchActions
                                    onChangesetPlanBatchActionClick={onChangesetPlanBatchActionClick}
                                    checkProvider={checkProvider}
                                    parsedQuery={parsedQuery}
                                    extensionsController={extensionsController}
                                />
                            </div>
                        </div>
                        {filteredDiagnosticsOrError.length === 0 ? (
                            <div className="container">
                                <p className="p-2 mb-0 text-muted">No diagnostics found.</p>
                            </div>
                        ) : (
                            <ul className="list-group list-group-flush mb-0">
                                {filteredDiagnosticsOrError.map((diagnostic, i) => (
                                    <li key={i} className={`list-group-item px-0 ${i === 0 ? 'border-top-0' : ''}`}>
                                        <DiagnosticsListItem
                                            {...props}
                                            key={JSON.stringify(diagnostic)}
                                            diagnostic={diagnostic}
                                            selectedAction={getSelectedActionForDiagnostic(diagnostic)}
                                            onActionSelect={onActionSelect}
                                            className="container-fluid"
                                            extensionsController={extensionsController}
                                        />
                                    </li>
                                ))}
                            </ul>
                        )}
                    </>
                )}
            </div>
        )
    }
)
