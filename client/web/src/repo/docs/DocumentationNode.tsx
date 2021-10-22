import classNames from 'classnames'
import * as H from 'history'
import BookOpenVariantIcon from 'mdi-react/BookOpenVariantIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import LinkVariantIcon from 'mdi-react/LinkVariantIcon'
import React, { RefObject, useEffect, useMemo, useRef } from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'

import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { AnchorLink } from '@sourcegraph/shared/src/components/Link'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { Badge } from '../../components/Badge'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepositoryFields } from '../../graphql-operations'
import { toDocumentationSingleSymbolURL, toDocumentationURL } from '../../util/url'

import { DocumentationExamples } from './DocumentationExamples'
import { DocumentationIcons } from './DocumentationIcons'
import { GQLDocumentationNode, Tag, isExcluded } from './graphql'
import { hasDescendent } from './RepositoryDocumentationSidebar'

interface Props extends Partial<RevisionSpec>, ResolvedRevisionSpec, BreadcrumbSetters, SettingsCascadeProps {
    repo: RepositoryFields

    history: H.History
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    commitID: string

    /** The documentation node to render */
    node: GQLDocumentationNode

    /** How far deep we are in the tree of documentation nodes */
    depth: number

    /** Whether or not this is the first child of a parent node */
    isFirstChild: boolean

    /** The pathID of the page containing this documentation node */
    pagePathID: string

    /** If specified, only render the documentation node corresponding to this path ID. e.g. "render just this one symbol" */
    onlyPathID?: string

    /** A list of documentation tags, a section will not be rendered if it matches one of these. */
    excludingTags: Tag[]

    /** The root scrolling area that this documentation node lives in. */
    scrollingRoot: RefObject<HTMLElement | undefined>

    /**
     * Called when the visibility of this documentation node changes.
     */
    onVisible: (node: GQLDocumentationNode, entry?: IntersectionObserverEntry) => void
}

export const DocumentationNode: React.FunctionComponent<Props> = React.memo(
    ({ useBreadcrumb, node, depth, isFirstChild, onlyPathID, scrollingRoot, onVisible, ...props }) => {
        const repoRevision = {
            repoName: props.repo.name,
            revision: props.revision || '',
        }
        const hashIndex = node.pathID.indexOf('#')
        const hash = hashIndex !== -1 ? node.pathID.slice(hashIndex + '#'.length) : ''
        let path = hashIndex !== -1 ? node.pathID.slice(0, hashIndex) : node.pathID
        path = path === '/' ? '' : path
        const thisPage = toDocumentationURL({ ...repoRevision, pathID: path + '#' + hash })
        const singleSymbolPage = toDocumentationSingleSymbolURL({ ...repoRevision, pathID: path + '#' + hash })

        useBreadcrumb(
            useMemo(
                () =>
                    depth === 0
                        ? { key: `node-${depth}`, element: <Link to={thisPage}>{node.label.value}</Link> }
                        : null,
                [depth, node.label.value, thisPage]
            )
        )

        const reference = useRef<HTMLDivElement>(null)
        const intersectionObserver = new IntersectionObserver(
            ([entry]) => {
                onVisible(node, entry)
            },
            {
                root: scrollingRoot?.current,
                threshold: [0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1],
            }
        )
        useEffect(() => {
            if (reference.current) {
                intersectionObserver.observe(reference.current)
            }
            // Remove the observer as soon as the component is unmounted
            return () => {
                onVisible(node)
                intersectionObserver.disconnect()
            }
        })

        if (node.detail.value === '') {
            const children = node.children.filter(child =>
                !child.node ? false : !isExcluded(child.node, props.excludingTags)
            )
            if (children.length === 0) {
                return null
            }
        }

        const headingLevel = depth + 1 < 4 ? depth + 1 : 4
        const topMargin =
            depth === 0
                ? 'mt-3' // Level 0 header ("Package foo")
                : onlyPathID
                ? 'mt-3' // Single-node display margin
                : depth === 1
                ? 'mt-5' // Level 1 headers ("Constants", "Variables", etc.)
                : isFirstChild
                ? 'mt-4'
                : 'mt-5' // Lowest level headers

        if (onlyPathID && node.pathID !== onlyPathID && !hasDescendent(node, onlyPathID)) {
            return null
        }
        const renderContent = !onlyPathID || node.pathID === onlyPathID || depth === 0
        return (
            <div className={classNames('documentation-node mb-5', topMargin)}>
                {renderContent && (
                    <div ref={reference}>
                        <Heading level={headingLevel} className="d-flex align-items-center documentation-node__heading">
                            <AnchorLink className="documentation-node__heading-anchor-link" to={thisPage}>
                                <LinkVariantIcon className="icon-inline" />
                            </AnchorLink>
                            {depth !== 0 && <DocumentationIcons className="mr-1" tags={node.documentation.tags} />}
                            <Link className="h" id={hash} to={singleSymbolPage}>
                                {node.label.value}
                            </Link>
                        </Heading>
                        {depth === 0 && (
                            <>
                                <div className="d-flex align-items-center mb-3">
                                    <span className="documentation-node__pill d-flex justify-content-center align-items-center px-2">
                                        <BookOpenVariantIcon className="icon-inline text-muted mr-1" /> Generated API
                                        docs
                                        <span className="documentation-node__pill-divider mx-2" />
                                        <a
                                            // eslint-disable-next-line react/jsx-no-target-blank
                                            target="_blank"
                                            rel="noopener"
                                            href="https://docs.sourcegraph.com/code_intelligence/apidocs"
                                        >
                                            Learn more
                                        </a>
                                    </span>
                                    {/*
                            TODO(apidocs): add support for indicating time the API docs were updated
                            <span className="ml-2">Last updated 2 days ago</span>
                        */}
                                    <Badge status="experimental" className="text-uppercase ml-2" />
                                </div>
                                <hr />
                                {onlyPathID && depth === 0 && (
                                    <Link className="mb-3 mt-2 d-inline-flex" to={thisPage}>
                                        ← View all of {node.label.value.toLowerCase()}
                                    </Link>
                                )}
                            </>
                        )}
                        {(!onlyPathID || node.pathID === onlyPathID) && node.detail.value !== '' && (
                            <div className="pt-2">
                                <Markdown dangerousInnerHTML={renderMarkdown(node.detail.value)} />
                            </div>
                        )}

                        {!isExcluded(node, ['test', 'benchmark', 'example', 'license', 'owner', 'package']) &&
                            node.documentation.tags.length !== 0 && (
                                <>
                                    <h4 className="mt-4">
                                        Usage examples
                                        <HelpCircleOutlineIcon
                                            className="icon-inline ml-1"
                                            data-tooltip="Usage examples from precise LSIF code intelligence index"
                                        />
                                    </h4>
                                    <DocumentationExamples {...props} pathID={node.pathID} count={onlyPathID ? 3 : 1} />
                                </>
                            )}
                    </div>
                )}

                {node.children?.map(
                    (child, index) =>
                        child.node &&
                        !isExcluded(child.node, props.excludingTags) && (
                            <DocumentationNode
                                key={`${depth}-${child.node.pathID}`}
                                {...props}
                                node={child.node}
                                depth={renderContent ? depth + 1 : depth}
                                isFirstChild={index === 0}
                                onlyPathID={
                                    onlyPathID ? (node.pathID === onlyPathID ? undefined : onlyPathID) : undefined
                                }
                                useBreadcrumb={useBreadcrumb}
                                scrollingRoot={scrollingRoot}
                                onVisible={onVisible}
                            />
                        )
                )}
            </div>
        )
    }
)

const Heading: React.FunctionComponent<{
    level: number
    children: React.ReactNode
    [x: string]: any
}> = ({ level, children, ...props }) => React.createElement(`h${level}`, props, children)
