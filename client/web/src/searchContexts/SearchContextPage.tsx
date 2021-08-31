import classNames from 'classnames'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useMemo } from 'react'
import { RouteComponentProps, useHistory, useLocation } from 'react-router'
import { catchError, startWith } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import { VersionContext } from '../schema/site.schema'
import {
    CaseSensitivityProps,
    ParsedSearchQueryProps,
    PatternTypeProps,
    SearchContextInputProps,
    SearchContextProps,
} from '../search'
import { SearchPageInput } from '../search/home/SearchPageInput'
import { convertMarkdownToBlocks } from '../search/notebook/convertMarkdownToBlocks'
import { RenderedSearchNotebookMarkdown } from '../search/notebook/RenderedSearchNotebookMarkdown'
import { SearchNotebookProps } from '../search/notebook/SearchNotebook'
import { ThemePreferenceProps } from '../theme'

import styles from './SearchContextPage.module.scss'

export interface SearchContextPageProps
    extends Pick<RouteComponentProps<{ spec: Scalars['ID'] }>, 'match'>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL'>,
        VersionContextProps,
        SearchContextInputProps,
        Pick<SearchContextProps, 'fetchSearchContextBySpec'>,
        Omit<SearchNotebookProps, 'onSerializeBlocks' | 'blocks' | 'collapseMenu'> {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined
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

    const useSearchNotebookDescription = useMemo(() => {
        if (!searchContextOrError || isErrorLike(searchContextOrError) || searchContextOrError === LOADING) {
            return false
        }
        const blocks = convertMarkdownToBlocks(searchContextOrError.description)
        return blocks.length > 1
    }, [searchContextOrError])

    const location = useLocation()
    const history = useHistory()

    return (
        <div className="w-100">
            <Page>
                <div className="container col-10">
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
                            <div className="my-3">
                                <SearchPageInput
                                    {...props}
                                    location={location}
                                    history={history}
                                    hideVersionContexts={true}
                                    selectedSearchContextSpec={searchContextOrError.spec}
                                    showOnboardingTour={false}
                                    showQueryBuilder={false}
                                    source="searchContextPage"
                                />
                            </div>
                            {searchContextOrError.description.length > 0 && (
                                <div className="my-3">
                                    {useSearchNotebookDescription ? (
                                        <div className="card px-3 pb-3">
                                            <RenderedSearchNotebookMarkdown
                                                {...props}
                                                markdown={searchContextOrError.description}
                                                collapseMenu={true}
                                            />
                                        </div>
                                    ) : (
                                        <div className="card p-3">
                                            <Markdown
                                                dangerousInnerHTML={renderMarkdown(searchContextOrError.description)}
                                            />
                                        </div>
                                    )}
                                </div>
                            )}
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
