import React, { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { DiagnosticInfo, getDiagnosticInfos } from '../../../threads/detail/backend'
const LOADING: 'loading' = 'loading'

/**
 * React component props for children of {@link WithChecklistsQueryResults}.
 */
export interface ChecklistsQueryResultProps {
    /** The list of checklists, loading, or an error. */
    checklistsOrError: typeof LOADING | DiagnosticInfo[] | ErrorLike
}

interface Props extends Partial<Pick<QueryParameterProps, 'query'>>, ExtensionsControllerProps {
    children: (props: ChecklistsQueryResultProps) => JSX.Element | null
}

/**
 * Wraps a component and provides a list of checklists resulting from querying using the provided
 * `query` prop.
 */
export const WithChecklistsQueryResults: React.FunctionComponent<Props> = ({ query, children, extensionsController }) => {
    const [diagnosticsOrError, setDiagnosticsOrError] = useState<typeof LOADING | DiagnosticInfo[] | ErrorLike>(LOADING)
    // tslint:disable-next-line: no-floating-promises
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            getDiagnosticInfos(extensionsController)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setDiagnosticsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [query, extensionsController])

    return children({ checklistsOrError: diagnosticsOrError })
}
