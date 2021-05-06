import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React from 'react'
import { RouteComponentProps } from 'react-router'
import { catchError, startWith } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { SearchContextProps } from '../search'

export interface SearchContextPageProps
    extends RouteComponentProps<{ id: Scalars['ID'] }>,
        Pick<SearchContextProps, 'fetchSearchContext'> {}

export const SearchContextPage: React.FunctionComponent<SearchContextPageProps> = props => {
    const LOADING = 'loading' as const

    const { match, history, fetchSearchContext } = props

    const searchContextOrError = useObservable(
        React.useMemo(
            () =>
                fetchSearchContext(match.params.id).pipe(
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [match.params.id, fetchSearchContext]
        )
    )

    return (
        <div className="w-100">
            <Page>
                <div className="container col-8">
                    {searchContextOrError === LOADING && <LoadingSpinner />}
                    {searchContextOrError && !isErrorLike(searchContextOrError) && searchContextOrError !== LOADING && (
                        <>
                            <PageTitle title={searchContextOrError.spec} />
                            <div className="mb-2 d-flex align-items-center">
                                <Link to="/contexts">
                                    <ChevronLeftIcon className="search-context-page__back-icon" />
                                </Link>
                                <h1 className="mb-0">{searchContextOrError.spec}</h1>
                                {!searchContextOrError.public && (
                                    <div className="badge badge-pill badge-secondary ml-1">Private</div>
                                )}
                            </div>
                            <div className="text-muted">
                                <span className="mr-1">{searchContextOrError.repositories.length} repositories</span>
                                &middot;
                                <span className="ml-1">
                                    Updated <Timestamp date={searchContextOrError.updatedAt} noAbout={true} />
                                </span>
                            </div>
                            <div className="my-2">
                                <Markdown
                                    dangerousInnerHTML={renderMarkdown(searchContextOrError.description)}
                                    history={history}
                                />
                            </div>
                            {!searchContextOrError.autoDefined && (
                                <>
                                    <div className="row mt-4">
                                        <div className="col">Repositories</div>
                                        <div className="col">Revisions</div>
                                    </div>
                                    <hr className="mt-2 mb-0" />
                                    {searchContextOrError.repositories.map(repositoryRevisions => (
                                        <div
                                            key={repositoryRevisions.repository.name}
                                            className="row search-context-page__repo-revs-row"
                                        >
                                            <div className="col search-context-page__repo-revs-row-repo">
                                                <Link to={`/${repositoryRevisions.repository.name}`}>
                                                    {repositoryRevisions.repository.name}
                                                </Link>
                                            </div>
                                            <div className="col">
                                                {repositoryRevisions.revisions.map(revision => (
                                                    <div
                                                        key={`${repositoryRevisions.repository.name}-${revision}`}
                                                        className="search-context-page__repo-revs-row-rev"
                                                    >
                                                        <Link
                                                            to={`/${repositoryRevisions.repository.name}@${revision}`}
                                                        >
                                                            {revision}
                                                        </Link>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    ))}
                                </>
                            )}
                        </>
                    )}
                    {isErrorLike(searchContextOrError) && (
                        <div className="alert alert-danger">
                            Error while loading the search context: <strong>{searchContextOrError.message}</strong>
                        </div>
                    )}
                </div>
            </Page>
        </div>
    )
}
