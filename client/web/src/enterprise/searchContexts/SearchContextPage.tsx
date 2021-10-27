import classNames from 'classnames'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import React from 'react'
import { RouteComponentProps } from 'react-router'
import { catchError, startWith } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { ISearchContextRepositoryRevisions } from '@sourcegraph/shared/src/graphql/schema'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { SearchContextProps } from '../../search'

import styles from './SearchContextPage.module.scss'

export interface SearchContextPageProps
    extends Pick<RouteComponentProps<{ spec: Scalars['ID'] }>, 'match'>,
        Pick<SearchContextProps, 'fetchSearchContextBySpec'> {}

const SearchContextRepositories: React.FunctionComponent<{ repositories: ISearchContextRepositoryRevisions[] }> = ({
    repositories,
}) => (
    <>
        <div className="d-flex">
            <div className="w-50">Repositories</div>
            <div className="w-50">Revisions</div>
        </div>
        <hr className="mt-2 mb-0" />
        {repositories.map(repositoryRevisions => (
            <div
                key={repositoryRevisions.repository.name}
                className={classNames(styles.searchContextPageRepoRevsRow, 'd-flex')}
            >
                <div className={classNames(styles.searchContextPageRepoRevsRowRepo, 'w-50')}>
                    <Link to={`/${repositoryRevisions.repository.name}`}>{repositoryRevisions.repository.name}</Link>
                </div>
                <div className="w-50">
                    {repositoryRevisions.revisions.map(revision => (
                        <div
                            key={`${repositoryRevisions.repository.name}-${revision}`}
                            className={styles.searchContextPageRepoRevsRowRev}
                        >
                            <Link to={`/${repositoryRevisions.repository.name}@${revision}`}>{revision}</Link>
                        </div>
                    ))}
                </div>
            </div>
        ))}
    </>
)

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
                                        icon: MagnifyIcon,
                                        to: '/search',
                                    },
                                    {
                                        to: '/contexts',
                                        text: 'Contexts',
                                    },
                                    {
                                        text: (
                                            <div className="d-flex align-items-center">
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
                                    <span className="ml-1">
                                        Updated <Timestamp date={searchContextOrError.updatedAt} noAbout={true} />
                                    </span>
                                </div>
                            )}
                            <Container className="mt-4">
                                {searchContextOrError.description && (
                                    <div className="mb-3">
                                        <Markdown
                                            dangerousInnerHTML={renderMarkdown(searchContextOrError.description)}
                                        />
                                    </div>
                                )}
                                {!searchContextOrError.autoDefined && (
                                    <h3>
                                        {searchContextOrError.repositories.length}{' '}
                                        {pluralize(
                                            'repository',
                                            searchContextOrError.repositories.length,
                                            'repositories'
                                        )}
                                    </h3>
                                )}
                                {!searchContextOrError.autoDefined && searchContextOrError.repositories.length > 0 && (
                                    <SearchContextRepositories repositories={searchContextOrError.repositories} />
                                )}
                            </Container>
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
