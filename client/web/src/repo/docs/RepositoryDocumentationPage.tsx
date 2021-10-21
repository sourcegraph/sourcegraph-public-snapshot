import classNames from 'classnames'
import * as H from 'history'
import { upperFirst } from 'lodash'
import BookOpenVariantIcon from 'mdi-react/BookOpenVariantIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useEffect, useCallback, useMemo, useState, useRef } from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError, ErrorLike } from '@sourcegraph/shared/src/util/errors'
import { RevisionSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container } from '@sourcegraph/wildcard'

import { Badge } from '../../components/Badge'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'
import { RepositoryFields } from '../../graphql-operations'
import { FeedbackPrompt } from '../../nav/Feedback/FeedbackPrompt'
import { routes } from '../../routes'
import { eventLogger } from '../../tracking/eventLogger'
import { toDocumentationURL } from '../../util/url'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'

import { DocumentationNode } from './DocumentationNode'
import { DocumentationWelcomeAlert } from './DocumentationWelcomeAlert'
import { fetchDocumentationPage, fetchDocumentationPathInfo, GQLDocumentationNode, isExcluded, Tag } from './graphql'
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
        BreadcrumbSetters,
        SettingsCascadeProps {
    repo: RepositoryFields
    history: H.History
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    pathID: string
    commitID: string
}

const LOADING = 'loading' as const

/** A page that shows a repository's documentation at the current revision. */
export const RepositoryDocumentationPage: React.FunctionComponent<Props> = React.memo(function Render({
    useBreadcrumb,
    ...props
}) {
    useEffect(() => {
        eventLogger.logViewEvent('RepositoryDocs')
    }, [])
    useScrollToLocationHash(props.location)

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
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [props.repo.id, props.commitID, pagePathID]
            )
        ) || LOADING

    const pathInfo =
        useObservable(
            useMemo(
                () =>
                    fetchDocumentationPathInfo({
                        repo: props.repo.id,
                        revspec: props.commitID,
                        pathID: pagePathID,
                        ignoreIndex: true,
                        maxDepth: 1,
                    }).pipe(
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [props.repo.id, props.commitID, pagePathID]
            )
        ) || LOADING

    const [sidebarVisible, setSidebarVisible] = useState(getSidebarVisibility())
    const handleSidebarVisible = useCallback((visible: boolean) => setSidebarVisible(visible), [])

    const loading = page === LOADING || pathInfo === LOADING
    const error = isErrorLike(page) ? page : isErrorLike(pathInfo) ? pathInfo : null

    const excludingTags: Tag[] = useMemo(() => ['private'], [])

    const containerReference = useRef<HTMLDivElement>(null)

    // Keep track of which node on the page is most visible, so that when visibility changes we can
    // know the active node and can apply various visual effects (like scrolling to it in the
    // sidebar.)
    const [visiblePathID, setVisiblePathID] = useState<string | null>(null)
    const [, setVisibilityEvents] = useState<{ pathID: string; intersectionRatio: number; element: HTMLElement }[]>([])
    const onVisible = React.useMemo(
        // eslint-disable-next-line unicorn/consistent-function-scoping
        () => (node: GQLDocumentationNode, entry?: IntersectionObserverEntry): void =>
            setVisibilityEvents(visibilityEvents => {
                // Update the list of currently-visible nodes.
                if (!entry || !entry.isIntersecting) {
                    // Remove all events for the now non-visible node.
                    visibilityEvents = visibilityEvents.filter(event => event.pathID !== node.pathID)
                } else {
                    // Add the new event.
                    visibilityEvents = visibilityEvents.filter(event => event.pathID !== node.pathID)
                    visibilityEvents.push({
                        pathID: node.pathID,
                        intersectionRatio: entry.intersectionRatio,
                        element: entry.target as HTMLElement,
                    })
                }

                if (containerReference.current) {
                    // Verify visibility of elements ourselves, because the IntersectionObserver API
                    // sometimes loses track of elements (does not fire a isIntersecting=false event)
                    // when scrolling very fast. I think that the IntersectionObserver v2 API solves
                    // this using the trackVisibility option, but we cannot use it except in Chrome:
                    // https://caniuse.com/intersectionobserver-v2
                    visibilityEvents = visibilityEvents.filter(event =>
                        isElementInView(event.element, containerReference.current!, true)
                    )

                    // Sort events by distance to the center of the screen. This way the "visible" node
                    // is always what's in the middle of your screen.
                    visibilityEvents.sort((a, b) => {
                        const aDistance = distanceToCenter(a.element, containerReference.current!)
                        const bDistance = distanceToCenter(b.element, containerReference.current!)
                        return aDistance < bDistance ? -1 : 1
                    })
                }
                if (visibilityEvents.length > 0) {
                    setVisiblePathID(visibilityEvents[0].pathID)
                }
                return visibilityEvents
            }),
        [setVisiblePathID, setVisibilityEvents]
    )

    // If we switch from rendering the entire page to rendering a specific full path ID (section of
    // the page), then scroll back to the top of the page as our scroll position would no longer be
    // meaningful.
    const onlyPathID = location.search === '' ? undefined : props.pathID + '#' + location.search.slice('?'.length)
    useEffect(() => {
        if (onlyPathID && containerReference.current) {
            containerReference.current.scrollTop = 0
        }
    }, [onlyPathID])

    return (
        <div className="repository-docs-page">
            {page !== LOADING && !isErrorLike(page) ? (
                <PageTitle
                    title={
                        onlyPathID
                            ? `${
                                  findDocumentationNode(page.tree, onlyPathID)?.documentation.searchKey ||
                                  page.tree.documentation.searchKey
                              } - ${displayRepoName(props.repo.name)} API docs`
                            : `${page.tree.documentation.searchKey} - ${displayRepoName(props.repo.name)} API docs`
                    }
                />
            ) : null}
            {loading ? <LoadingSpinner className="icon-inline m-1" /> : null}
            {error && error.message === 'page not found' ? <PageNotFound /> : null}
            {error && (error.message === 'no LSIF data' || error.message === 'no LSIF documentation') ? (
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
                            <h2 className="text-muted mb-2">
                                <MapSearchIcon className="icon-inline mr-2" />
                                Repository has no LSIF documentation data
                            </h2>
                            <p className="mt-3">
                                Sourcegraph can use LSIF code intelligence to generate API documentation for all your
                                code, giving you the ability to navigate and explore the APIs provided by this
                                repository.
                            </p>
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
                            <p className="text-muted mt-3 mb-0">
                                <strong>Note:</strong> only the Go programming language is currently supported.
                            </p>
                        </Container>
                    </div>
                </div>
            ) : null}
            {isErrorLike(error) &&
            error.message !== 'page not found' &&
            error.message !== 'no LSIF data' &&
            error.message !== 'no LSIF documentation' ? (
                <PageError error={error} />
            ) : null}
            {page !== LOADING && !isErrorLike(page) && pathInfo !== LOADING && !isErrorLike(pathInfo) ? (
                <>
                    <RepositoryDocumentationSidebar
                        {...props}
                        onToggle={handleSidebarVisible}
                        node={page.tree}
                        pathInfo={pathInfo}
                        pagePathID={pagePathID}
                        activePathID={visiblePathID || pagePathID}
                        depth={0}
                    />
                    <div className="repository-docs-page__container" ref={containerReference}>
                        <div
                            className={classNames(
                                'repository-docs-page__container-content',
                                sidebarVisible && 'repository-docs-page__container-content--sidebar-visible'
                            )}
                        >
                            {/*
                                TODO(apidocs): Eventually this welcome alert should go away entirely, but for now
                                it's the best thing we have for the sometimes empty root landing page.
                            */}
                            {page.tree.detail.value === '' && <DocumentationWelcomeAlert />}
                            {isExcluded(page.tree, excludingTags) ? (
                                <div className="m-3">
                                    <h2 className="text-muted">Looks like there's nothing to see here.</h2>
                                    <p>API docs for private / unexported code is coming soon!</p>
                                </div>
                            ) : null}
                            <DocumentationNode
                                {...props}
                                useBreadcrumb={useBreadcrumb}
                                node={page.tree}
                                pagePathID={pagePathID}
                                depth={0}
                                isFirstChild={true}
                                onlyPathID={onlyPathID}
                                excludingTags={excludingTags}
                                scrollingRoot={containerReference}
                                onVisible={onVisible}
                            />
                        </div>
                    </div>
                </>
            ) : null}
        </div>
    )
})

/** Finds a descendant child node of the input with the given path ID. */
function findDocumentationNode(node: GQLDocumentationNode, pathID: string): GQLDocumentationNode | undefined {
    if (node.pathID === pathID) {
        return node
    }
    for (const child of node.children) {
        if (child.node) {
            const found = findDocumentationNode(child.node, pathID)
            if (found) {
                return found
            }
        }
    }
    return undefined
}

/** Checks if an element is in view of the scrolling container. */
function isElementInView(element: HTMLElement, container: HTMLElement, partial: boolean): boolean {
    const containerTop = container.scrollTop
    const containerBottom = containerTop + container.clientHeight

    const elementTop = element.offsetTop
    const elementBottom = elementTop + element.clientHeight

    if (elementTop >= containerTop && elementBottom <= containerBottom) {
        return true
    }
    return (
        (partial && elementTop < containerTop && elementBottom > containerTop) ||
        (elementBottom > containerBottom && elementTop < containerBottom)
    )
}

/**
 * Returns the distance between the element's area (whichever point is lesser) and the scrolling
 * container's viewport center. i.e., how far away the element is from being in the middle of the
 * scrolling container's viewport.
 */
function distanceToCenter(element: HTMLElement, container: HTMLElement): number {
    const containerTop = container.scrollTop
    const containerBottom = containerTop + container.clientHeight
    const containerHeight = containerBottom - containerTop
    const containerCenter = containerTop + containerHeight / 2

    const elementTop = element.offsetTop
    const elementBottom = elementTop + element.clientHeight
    const elementHeight = elementBottom - elementTop
    const elementCenter = elementTop + elementHeight / 2

    if (elementTop < containerCenter && elementBottom > containerCenter) {
        return 0
    }
    return absolute(containerCenter - elementCenter)
}

function absolute(value: number): number {
    return value < 0 ? -value : value
}
