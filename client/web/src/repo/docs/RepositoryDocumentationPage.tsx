import * as H from 'history'
import { upperFirst } from 'lodash'
import BookOpenVariantIcon from 'mdi-react/BookOpenVariantIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useEffect, useCallback, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { catchError, map, startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError, ErrorLike } from '@sourcegraph/shared/src/util/errors'
import { RevisionSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container } from '@sourcegraph/wildcard'

import { Badge } from '../../components/Badge'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { RepositoryFields } from '../../graphql-operations'
import { FeedbackPrompt } from '../../nav/Feedback/FeedbackPrompt'
import { routes } from '../../routes'
import { eventLogger } from '../../tracking/eventLogger'
import { toDocumentationURL } from '../../util/url'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'

import { DocumentationNode } from './DocumentationNode'
import { DocumentationWelcomeAlert } from './DocumentationWelcomeAlert'
import { RepositoryDocumentationSidebar, getSidebarVisibility } from './RepositoryDocumentationSidebar'

const PageError: React.FunctionComponent<{ error: ErrorLike }> = ({ error }) => (
    <div className="repository-docs-page__error alert alert-danger m-2">Error: {upperFirst(error.message)}</div>
)

const PageNotFound: React.FunctionComponent = () => (
    <div className="repository-docs-page__not-found">
        <MapSearchIcon className="icon-inline" /> Page not found
    </div>
)

interface Props
    extends RepoHeaderContributionsLifecycleProps,
        Partial<RevisionSpec>,
        ResolvedRevisionSpec,
        BreadcrumbSetters {
    repo: RepositoryFields
    history: H.History
    location: H.Location
    pathID: string
    commitID: string
}

const LOADING = 'loading' as const

/** A page that shows a repository's documentation at the current revision. */
export const RepositoryDocumentationPage: React.FunctionComponent<Props> = ({ useBreadcrumb, ...props }) => {
    // TODO(slimsag): nightmare: there is _something_ in the props that causes this entire page to
    // rerender whenever you type in the search bar. In fact, this also appears to happen on all other
    // pages!
    //
    // 1. Navigate to https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/compare/HEAD~40...HEAD?visible=45
    // 2. Scroll to the bottom of this page twice and click "Show more"
    // 3. Try typing in the search bar at the top of the page and the whole thing stutters a ton
    //
    // See https://sourcegraph.slack.com/archives/C01C3NCGD40/p1621485604017100
    // and https://github.com/sourcegraph/sourcegraph/issues/21200
    useEffect(() => {
        eventLogger.logViewEvent('RepositoryDocs')
    }, [])

    const thisPage = toDocumentationURL({ repoName: props.repo.name, revision: props.revision || '', pathID: '' })
    useBreadcrumb(useMemo(() => ({ key: 'node', element: <Link to={thisPage}>API docs</Link> }), [thisPage]))

    const pagePathID = props.pathID || '/'
    const page =
        useObservable(
            useMemo(
                () =>
                    fetchDocumentationPage({
                        repo: props.repo.id,
                        revspec: props.commitID,
                        pathID: pagePathID,
                    }).pipe(
                        map(page => ({ ...page, tree: JSON.parse(page.tree) as GQLDocumentationNode })),
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [props.repo.id, props.commitID, pagePathID]
            )
        ) || LOADING

    const [sidebarVisible, setSidebarVisible] = useState(getSidebarVisibility())
    const handleSidebarVisible = useCallback((visible: boolean) => setSidebarVisible(visible), [])

    return (
        <div className="repository-docs-page">
            <PageTitle title="API docs" />
            {page === LOADING ? <LoadingSpinner className="icon-inline m-1" /> : null}
            {isErrorLike(page) && page.message === 'page not found' ? <PageNotFound /> : null}
            {isErrorLike(page) && (page.message === 'no LSIF data' || page.message === 'no LSIF documentation') ? (
                <div className="repository-docs-page__container">
                    <div className="repository-docs-page__container-content">
                        <div className="d-flex float-right">
                            <a
                                // eslint-disable-next-line react/jsx-no-target-blank
                                target="_blank"
                                rel="noopener"
                                href="https://docs.sourcegraph.com/code_intelligence/apidocs"
                                className="mr-1 btn btn-sm text-decoration-none btn-link btn-outline-secondary"
                            >
                                Learn more
                            </a>
                            <FeedbackPrompt routes={routes} />
                        </div>
                        <h1>
                            <BookOpenVariantIcon className="icon-inline mr-1" />
                            API docs
                            <Badge status="experimental" className="text-uppercase ml-2" />
                        </h1>
                        <p>API documentation generated for all your code</p>
                        <Container>
                            <h2 className="text-muted mb-2">This repository has no LSIF documentation data.</h2>
                            <h3>
                                <a
                                    // eslint-disable-next-line react/jsx-no-target-blank
                                    target="_blank"
                                    rel="noopener"
                                    href="https://docs.sourcegraph.com/code_intelligence/apidocs"
                                >
                                    Learn more
                                </a>
                            </h3>
                        </Container>
                    </div>
                </div>
            ) : null}
            {isErrorLike(page) &&
            page.message !== 'page not found' &&
            page.message !== 'no LSIF data' &&
            page.message !== 'no LSIF documentation' ? (
                <PageError error={page} />
            ) : null}
            {page !== LOADING && !isErrorLike(page) ? (
                <>
                    <RepositoryDocumentationSidebar
                        {...props}
                        onToggle={handleSidebarVisible}
                        node={page.tree}
                        pagePathID={pagePathID}
                        depth={0}
                    />
                    <div className="repository-docs-page__container">
                        <div
                            className={`repository-docs-page__container-content${
                                sidebarVisible ? ' repository-docs-page__container-content--sidebar-visible' : ''
                            }`}
                        >
                            <DocumentationWelcomeAlert />
                            <DocumentationNode
                                {...props}
                                useBreadcrumb={useBreadcrumb}
                                node={page.tree}
                                pagePathID={pagePathID}
                                depth={0}
                                excludingTags={['private']}
                            />
                        </div>
                    </div>
                    <div className="repository-docs-page__feedback-container">
                        <div className="repository-docs-page__feedback-container-content">
                            <Badge status="experimental" className="text-uppercase mr-2" />
                            <a
                                // eslint-disable-next-line react/jsx-no-target-blank
                                target="_blank"
                                rel="noopener"
                                href="https://docs.sourcegraph.com/code_intelligence/apidocs"
                                className="mr-1 btn btn-sm text-decoration-none btn-link btn-outline-secondary"
                            >
                                Learn more
                            </a>
                            <FeedbackPrompt routes={routes} />
                        </div>
                    </div>
                </>
            ) : null}
        </div>
    )
}
