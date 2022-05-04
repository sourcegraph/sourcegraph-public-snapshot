import React, { useCallback, useMemo } from 'react'

import MagnifyIcon from 'mdi-react/MagnifyIcon'
import { RouteComponentProps } from 'react-router'
import { Observable, of, throwError } from 'rxjs'
import { catchError, startWith, switchMap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import {
    Scalars,
    SearchContextEditInput,
    SearchContextRepositoryRevisionsInput,
} from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ISearchContext } from '@sourcegraph/shared/src/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader, LoadingSpinner, useObservable, Alert } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'

import { SearchContextForm } from './SearchContextForm'

export interface EditSearchContextPageProps
    extends RouteComponentProps<{ spec: Scalars['ID'] }>,
        ThemeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'updateSearchContext' | 'fetchSearchContextBySpec' | 'deleteSearchContext'>,
        PlatformContextProps<'requestGraphQL'> {
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

export const AuthenticatedEditSearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<EditSearchContextPageProps>
> = props => {
    const LOADING = 'loading' as const

    const { match, updateSearchContext, fetchSearchContextBySpec, platformContext } = props
    const onSubmit = useCallback(
        (
            id: Scalars['ID'] | undefined,
            searchContext: SearchContextEditInput,
            repositories: SearchContextRepositoryRevisionsInput[]
        ): Observable<ISearchContext> => {
            if (!id) {
                return throwError(new Error('Cannot update search context with undefined ID'))
            }
            return updateSearchContext({ id, searchContext, repositories }, platformContext)
        },
        [updateSearchContext, platformContext]
    )

    const searchContextOrError = useObservable(
        useMemo(
            () =>
                fetchSearchContextBySpec(match.params.spec, platformContext).pipe(
                    switchMap(searchContext => {
                        if (!searchContext.viewerCanManage) {
                            return throwError(new Error('You do not have sufficient permissions to edit this context.'))
                        }
                        return of(searchContext)
                    }),
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [match.params.spec, fetchSearchContextBySpec, platformContext]
        )
    )

    return (
        <div className="w-100">
            <Page>
                <div className="container col-8">
                    <PageTitle title="Edit context" />
                    <PageHeader
                        className="mb-3"
                        path={[
                            {
                                icon: MagnifyIcon,
                                to: '/search',
                                ariaLabel: 'Code Search',
                            },
                            {
                                to: '/contexts',
                                text: 'Contexts',
                            },
                            {
                                text: 'Edit context',
                            },
                        ]}
                    />
                    {searchContextOrError === LOADING && (
                        <div className="d-flex justify-content-center">
                            <LoadingSpinner inline={false} />
                        </div>
                    )}
                    {searchContextOrError && searchContextOrError !== LOADING && !isErrorLike(searchContextOrError) && (
                        <SearchContextForm {...props} searchContext={searchContextOrError} onSubmit={onSubmit} />
                    )}
                    {isErrorLike(searchContextOrError) && (
                        <Alert data-testid="search-contexts-alert-danger" variant="danger">
                            Error while loading the search context: <strong>{searchContextOrError.message}</strong>
                        </Alert>
                    )}
                </div>
            </Page>
        </div>
    )
}

export const EditSearchContextPage = withAuthenticatedUser(AuthenticatedEditSearchContextPage)
