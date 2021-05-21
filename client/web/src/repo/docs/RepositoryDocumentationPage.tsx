import * as H from 'history'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useEffect, useCallback, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike } from '@sourcegraph/shared/src/util/errors'
import { RevisionSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { requestGraphQL } from '../../backend/graphql'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { RepositoryFields, Scalars } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { toDocumentationURL } from '../../util/url'

import { DocumentationNode, GQLDocumentationNode } from './DocumentationNode'
import { RepositoryDocumentationSidebar, getSidebarVisibility } from './RepositoryDocumentationSidebar'

interface DocumentationPageResults {
    node: GQL.IRepository
}
interface DocumentationPageVariables {
    repo: Scalars['ID']
    revspec: string
    pathID: string
}

const fetchDocumentationPage = (args: DocumentationPageVariables): Observable<GQL.IDocumentationPage> =>
    requestGraphQL<DocumentationPageResults, DocumentationPageVariables>(
        gql`
            query DocumentationPage($repo: ID!, $revspec: String!, $pathID: String!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $revspec) {
                            tree(path: "/") {
                                lsif {
                                    documentationPage(pathID: $pathID) {
                                        tree
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node) {
                throw createAggregateError(errors)
            }
            const repo = data.node as GQL.IRepository
            if (
                !repo.commit ||
                !repo.commit.tree ||
                !repo.commit.tree.lsif ||
                !repo.commit.tree.lsif.documentationPage ||
                !repo.commit.tree.lsif.documentationPage.tree
            ) {
                throw new Error('page not found')
            }
            return repo.commit.tree.lsif.documentationPage
        })
    )

export const PageError: React.FunctionComponent<{ error: ErrorLike }> = ({ error }) => (
    <div className="repository-docs-page__error alert alert-danger m-2">
        <AlertCircleIcon className="redesign-d-none icon-inline" /> Error: {upperFirst(error.message)}
    </div>
)

const PageNotFound: React.FunctionComponent = () => (
    <div className="repository-docs-page__not-found">
        <MapSearchIcon className="icon-inline" /> Page not found
    </div>
)

export interface RepositoryDocumentationPageProps
    extends Partial<RevisionSpec>,
        ResolvedRevisionSpec,
        BreadcrumbSetters {
    repo: RepositoryFields

    history: H.History
    location: H.Location
    pathID: string
}

const LOADING = 'loading' as const

/** A page that shows a repository's documentation at the current revision. */
export const RepositoryDocumentationPage: React.FunctionComponent<RepositoryDocumentationPageProps> = ({
    useBreadcrumb,
    ...props
}) => {
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
            {isErrorLike(page) && page.message !== 'page not found' ? <PageError error={page} /> : null}
            {page !== LOADING && !isErrorLike(page) ? (
                <>
                    <RepositoryDocumentationSidebar
                        {...props}
                        className="repository-docs-page__sidebar"
                        onToggle={handleSidebarVisible}
                        node={page.tree}
                        pagePathID={pagePathID}
                        depth={0}
                    />
                    <div
                        className={`repository-docs-page__container${
                            sidebarVisible ? ' repository-docs-page__container--sidebar-visible' : ''
                        }`}
                    >
                        <div className="repository-docs-page__container-content">
                            <DocumentationNode
                                {...props}
                                useBreadcrumb={useBreadcrumb}
                                node={page.tree}
                                pagePathID={pagePathID}
                                depth={0}
                            />
                        </div>
                    </div>
                </>
            ) : null}
        </div>
    )
}
