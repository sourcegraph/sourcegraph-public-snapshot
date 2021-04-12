import React from 'react'
import { RouteComponentProps } from 'react-router'
import { catchError, startWith } from 'rxjs/operators'

import { Scalars } from '../../../shared/src/graphql-operations'
import { asError, isErrorLike } from '../../../shared/src/util/errors'
import { useObservable } from '../../../shared/src/util/useObservable'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { fetchSearchContext } from '../search/backend'

export interface SearchContextPageProps extends RouteComponentProps<{ id: Scalars['ID'] }> {}

export const SearchContextPage: React.FunctionComponent<SearchContextPageProps> = props => {
    const LOADING = 'loading' as const

    const { match } = props

    const searchContextOrError = useObservable(
        React.useMemo(
            () =>
                fetchSearchContext(match.params.id).pipe(
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [match.params.id]
        )
    )

    return (
        <div className="w-100">
            <Page>
                {searchContextOrError === LOADING && 'Loading'}
                {searchContextOrError && !isErrorLike(searchContextOrError) && searchContextOrError !== LOADING && (
                    <>
                        <PageTitle title={searchContextOrError.spec} />
                        <h1>{searchContextOrError.spec}</h1>
                        <div className="mb-2">Description: {searchContextOrError.description}</div>
                        <ul className="list-group list-group-flush">
                            {searchContextOrError.repositories.map(repo => (
                                <li className="list-group-item mb-1" key={repo.repository.name}>
                                    {repo.repository.name}
                                    <ul>
                                        {repo.revisions.map(revision => (
                                            <li key={revision}>{revision}</li>
                                        ))}
                                    </ul>
                                </li>
                            ))}
                        </ul>
                    </>
                )}
                {isErrorLike(searchContextOrError) && (
                    <div className="alert alert-danger">
                        Error while loading the search context: <strong>{searchContextOrError.message}</strong>
                    </div>
                )}
            </Page>
        </div>
    )
}
