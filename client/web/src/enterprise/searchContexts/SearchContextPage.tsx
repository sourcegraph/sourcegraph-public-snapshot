import classNames from 'classnames'
import { debounce } from 'lodash'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { catchError, startWith } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import { Scalars, SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { ISearchContextRepositoryRevisions } from '@sourcegraph/shared/src/graphql/schema'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Badge, Container, PageHeader } from '@sourcegraph/wildcard'

import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'
import { SearchContextProps } from '../../search'

import styles from './SearchContextPage.module.scss'

export interface SearchContextPageProps
    extends Pick<RouteComponentProps<{ spec: Scalars['ID'] }>, 'match'>,
        Pick<SearchContextProps, 'fetchSearchContextBySpec'> {}

const initialRepositoriesToShow = 15
const incrementalRepositoriesToShow = 10

const SearchContextRepositories: React.FunctionComponent<{ repositories: ISearchContextRepositoryRevisions[] }> = ({
    repositories,
}) => {
    const [filterQuery, setFilterQuery] = useState('')
    const debouncedSetFilterQuery = useMemo(() => debounce(value => setFilterQuery(value), 250), [setFilterQuery])
    const filteredRepositories = useMemo(
        () =>
            repositories.filter(repositoryRevisions => {
                const lowerCaseFilterQuery = filterQuery.toLowerCase()
                return (
                    !lowerCaseFilterQuery ||
                    repositoryRevisions.repository.name.toLowerCase().includes(lowerCaseFilterQuery) ||
                    repositoryRevisions.revisions.some(revision =>
                        revision.toLowerCase().includes(lowerCaseFilterQuery)
                    )
                )
            }),
        [repositories, filterQuery]
    )

    const [repositoriesToShow, setRepositoriesToShow] = useState(initialRepositoriesToShow)
    const onBottomHit = useCallback(
        () =>
            setRepositoriesToShow(repositoriesToShow =>
                Math.min(filteredRepositories.length, repositoriesToShow + incrementalRepositoriesToShow)
            ),
        [filteredRepositories]
    )

    const renderRepositoryRevisions = useCallback(
        (repositoryRevisions: ISearchContextRepositoryRevisions) => (
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
        ),
        []
    )

    return (
        <>
            <div className="d-flex justify-content-between align-items-center mb-3">
                {filteredRepositories.length > 0 && (
                    <h3>
                        <span>
                            {filteredRepositories.length}{' '}
                            {pluralize('repository', filteredRepositories.length, 'repositories')}
                        </span>
                    </h3>
                )}
                {repositories.length > 0 && (
                    <input
                        type="text"
                        className="form-control form-control-md w-50"
                        placeholder="Search repositories and revisions"
                        onChange={event => debouncedSetFilterQuery(event.target.value)}
                    />
                )}
            </div>
            {repositories.length > 0 && (
                <>
                    <div className="d-flex">
                        <div className="w-50">Repositories</div>
                        <div className="w-50">Revisions</div>
                    </div>
                    <hr className="mt-2 mb-0" />
                    <VirtualList<ISearchContextRepositoryRevisions>
                        className="mt-2"
                        itemsToShow={repositoriesToShow}
                        onShowMoreItems={onBottomHit}
                        items={filteredRepositories}
                        itemProps={undefined}
                        itemKey={repositoryRevisions => repositoryRevisions.repository.name}
                        renderItem={renderRepositoryRevisions}
                    />
                </>
            )}
        </>
    )
}

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
                                                    <Badge
                                                        variant="secondary"
                                                        pill={true}
                                                        className={classNames(
                                                            'ml-2',
                                                            styles.searchContextPagePrivateBadge
                                                        )}
                                                        as="div"
                                                    >
                                                        Private
                                                    </Badge>
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
                                {!searchContextOrError.autoDefined && searchContextOrError.query.length === 0 && (
                                    <SearchContextRepositories repositories={searchContextOrError.repositories} />
                                )}
                                {searchContextOrError.query.length > 0 && (
                                    <Link
                                        to={`/search?${buildSearchURLQuery(
                                            searchContextOrError.query,
                                            SearchPatternType.regexp,
                                            false
                                        )}`}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >
                                        <SyntaxHighlightedSearchQuery query={searchContextOrError.query} />
                                    </Link>
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
