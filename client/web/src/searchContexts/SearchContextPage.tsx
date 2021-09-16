import classNames from 'classnames'
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
import { PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { SearchContextProps } from '../search'

import styles from './SearchContextPage.module.scss'

export interface SearchContextPageProps
    extends Pick<RouteComponentProps<{ spec: Scalars['ID'] }>, 'match'>,
        Pick<SearchContextProps, 'fetchSearchContextBySpec'> {}

export const SearchContextPage: React.FunctionComponent<SearchContextPageProps> = props => {
    const LOADING = 'loading' as const

    const { match, fetchSearchContextBySpec } = props

    const searchContextOrError = useObservable(
        React.useMemo(
            () =>
                fetchSearchContextBySpec(match.params.spec).pipe(
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [match.params.spec, fetchSearchContextBySpec]
        )
    )

    return (
        <div className="w-100">
            <Page>
                <div className="container col-8">
                    {searchContextOrError === LOADING && (
                        <div className="d-flex justify-content-center">
                            <LoadingSpinner />
                        </div>
                    )}
                    {searchContextOrError && !isErrorLike(searchContextOrError) && searchContextOrError !== LOADING && (
                        <>
                            <PageTitle title={searchContextOrError.spec} />
                            <PageHeader
                                className="mb-2"
                                path={[
                                    {
                                        text: (
                                            <div className="d-flex align-items-center">
                                                <Link
                                                    className="d-flex"
                                                    to="/contexts"
                                                    aria-label="Back to contexts list"
                                                    title="Back to contexts list"
                                                >
                                                    <ChevronLeftIcon />
                                                </Link>
                                                <span>{searchContextOrError.spec}</span>
                                                {!searchContextOrError.public && (
                                                    <div
                                                        className={classNames(
                                                            'badge badge-pill badge-secondary ml-2',
                                                            styles.searchContextPagePrivateBadge
                                                        )}
                                                    >
                                                        Private
                                                    </div>
                                                )}
                                            </div>
                                        ),
                                    },
                                ]}
                                actions={
                                    searchContextOrError.viewerCanManage && (
                                        <Link
                                            to={`/contexts/${searchContextOrError.spec}/edit`}
                                            className="btn btn-secondary"
                                            data-testid="edit-search-context-link"
                                        >
                                            Edit
                                        </Link>
                                    )
                                }
                            />
                            {!searchContextOrError.autoDefined && (
                                <div className="text-muted">
                                    <span className="mr-1">
                                        {searchContextOrError.repositories.length} repositories
                                    </span>
                                    &middot;
                                    <span className="ml-1">
                                        Updated <Timestamp date={searchContextOrError.updatedAt} noAbout={true} />
                                    </span>
                                </div>
                            )}
                            <div className="my-2">
                                <Markdown dangerousInnerHTML={renderMarkdown(searchContextOrError.description)} />
                            </div>
                            {!searchContextOrError.autoDefined && (
                                <>
                                    <div className="mt-4 d-flex">
                                        <div className="w-50">Repositories</div>
                                        <div className="w-50">Revisions</div>
                                    </div>
                                    <hr className="mt-2 mb-0" />
                                    {searchContextOrError.repositories.map(repositoryRevisions => (
                                        <div
                                            key={repositoryRevisions.repository.name}
                                            className={classNames(styles.searchContextPageRepoRevsRow, 'd-flex')}
                                        >
                                            <div
                                                className={classNames(styles.searchContextPageRepoRevsRowRepo, 'w-50')}
                                            >
                                                <Link to={`/${repositoryRevisions.repository.name}`}>
                                                    {repositoryRevisions.repository.name}
                                                </Link>
                                            </div>
                                            <div className="w-50">
                                                {repositoryRevisions.revisions.map(revision => (
                                                    <div
                                                        key={`${repositoryRevisions.repository.name}-${revision}`}
                                                        className={styles.searchContextPageRepoRevsRowRev}
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
