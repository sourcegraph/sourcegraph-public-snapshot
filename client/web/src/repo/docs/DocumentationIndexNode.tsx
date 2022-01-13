import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import CircleMediumIcon from 'mdi-react/CircleMediumIcon'
import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Button } from '@sourcegraph/wildcard'

import { RepositoryFields } from '../../graphql-operations'
import { toDocumentationURL } from '../../util/url'

import styles from './DocumentationIndexNode.module.scss'
import { DocumentationNodeChild, GQLDocumentationNode, isExcluded, Tag } from './graphql'

/**
 * Mirrors the GraphQL DocumentationNodeChild interface, but swaps the node out with an extended
 * IndexNode interface.
 */
export interface IndexNodeChild extends DocumentationNodeChild {
    node?: IndexNode
    pathID?: string
}

/**
 * Mirrors the GQLDocumentationNode interface, extending it with whether or not nodes are active
 * or lead to the active node.
 */
export interface IndexNode extends GQLDocumentationNode {
    /** Children of this node */
    children: IndexNodeChild[]

    /** Whether or not this node is currently active / being looked at on the screen. */
    isActive: boolean

    /** Whether or not this node is in the path of nodes leading to the currently active one. */
    inActivePath: boolean
}

interface Props extends Partial<RevisionSpec>, ResolvedRevisionSpec {
    repo: RepositoryFields

    history: H.History
    location: H.Location

    /** The documentation node to render */
    node: IndexNode

    /** How far deep we are in the tree of documentation nodes */
    depth: number

    /** The pathID of the page containing this documentation node */
    pagePathID: string

    /** A list of documentation tags, a section will not be rendered if it matches one of these. */
    excludingTags: Tag[]
}

export const DocumentationIndexNode: React.FunctionComponent<Props> = React.memo(
    ({ node, depth, ...props }) => {
        const repoRevision = {
            repoName: props.repo.name,
            revision: props.revision || '',
        }
        const hashIndex = node.pathID.indexOf('#')
        const hash = hashIndex !== -1 ? node.pathID.slice(hashIndex + '#'.length) : ''
        let path = hashIndex !== -1 ? node.pathID.slice(0, hashIndex) : node.pathID
        path = path === '/' ? '' : path
        const thisPage = toDocumentationURL({ ...repoRevision, pathID: path + '#' + hash })

        // Keep track of the expanded state the user has requested.
        const [userExpanded, setUserExpanded] = useState(false)

        // Keep track of the actual expanded state we will use.
        const autoExpand = depth === 0 || node.isActive || node.inActivePath
        const [expanded, setExpanded] = useState(autoExpand)
        const toggleExpanded = (): void => {
            setUserExpanded(expanded => !expanded)
            setExpanded(expanded => !expanded)
        }

        // If a new node has come into view, automatically expand - or if no longer in view, collapse.
        const numberChildren = node.children.length
        useEffect(() => {
            // If the user explicitly expanded, respect them and don't collapse.
            if (!userExpanded) {
                // Don't collapse an area we previously expanded unless there are a large number of
                // children in it, otherwise all the expanding/collapsing is a lot of moving and
                // too jarring.
                if (autoExpand || (!autoExpand && numberChildren > 30)) {
                    setExpanded(autoExpand)
                }
            }
        }, [autoExpand, userExpanded, numberChildren])

        // If this node becomes the active one (the one being viewed), scroll this index (sidebar) node
        // into view.
        const nodeReference = React.useRef<HTMLDivElement>(null)
        useEffect(() => {
            if (node.isActive) {
                if (depth === 0) {
                    setTimeout(() => {
                        if (nodeReference.current?.parentElement) {
                            nodeReference.current.parentElement.scrollTo({ top: 0, behavior: 'smooth' })
                        }
                    }, 250)
                } else {
                    nodeReference.current?.scrollIntoView({
                        /**
                         * It would be nice to use smooth scrolling here, in Firefox it's very nice,
                         * but unfortunately Chrome seems to do some debouncing of smooth scrolling
                         * which means the sidebar then doesn't follow as you scroll at all. Lame.
                         */
                        behavior: 'auto',
                        block: 'center',
                    })
                }
            }
        }, [node.isActive, depth, nodeReference])

        const excluded = isExcluded(node, props.excludingTags)
        if (excluded) {
            return null
        }

        if (node.detail.value === '') {
            const children = node.children.filter(child =>
                !child.node ? false : !isExcluded(child.node, props.excludingTags)
            )
            if (children.length === 0) {
                return null
            }
        }

        // We do make use of useScrollToLocationHash for navigating to the clicked sidebar node,
        // but the speed at which a browser URL change occurs, triggers a rerender and can be
        // picked up by useScrollToLocationHash is noticeably slow. To workaround this, we manually
        // scroll the element into view as soon as the link is clicked.
        const scrollToFast = (): void => {
            // eslint-disable-next-line unicorn/prefer-query-selector
            const element = document.getElementById(hash)
            if (element) {
                element.scrollIntoView()
            }
        }

        const styleAsActive = node.children.length === 0 && node.isActive
        const styleAsExpandable = !styleAsActive && depth !== 0 && node.children.length > 0
        return (
            <div
                className={classNames('documentation-index-node d-flex flex-column', depth !== 0 && 'mt-2')}
                ref={nodeReference}
            >
                <span
                    className={classNames(
                        'd-flex align-items-center text-nowrap',
                        styles.documentationIndexNodeRow,
                        (styleAsActive || styleAsExpandable) && styles.documentationIndexNodeRowShiftLeft
                    )}
                >
                    {styleAsActive && (
                        <CircleMediumIcon
                            className={classNames(
                                'd-flex flex-shrink-0 mr-1 icon-inline',
                                styles.documentationIndexNodeActiveCircle
                            )}
                        />
                    )}
                    {styleAsExpandable && (
                        <Button
                            className={classNames(
                                'd-flex flex-shrink-0 mr-1 btn-icon',
                                styles.documentationIndexNodeExpandButton
                            )}
                            aria-label={expanded ? 'Collapse section' : 'Expand section'}
                            onClick={toggleExpanded}
                        >
                            {expanded ? (
                                <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                            ) : (
                                <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                            )}
                            {node.detail.value === '' && <strong id={'index-' + hash}>{node.label.value}</strong>}
                        </Button>
                    )}
                    {node.detail.value !== '' && (
                        <Link id={'index-' + hash} to={thisPage} onClick={scrollToFast} className="pr-3">
                            {node.label.value}
                        </Link>
                    )}
                </span>
                {expanded && (
                    <ul className="pl-3">
                        {node.children?.map(child =>
                            child.pathID ? null : (
                                <DocumentationIndexNode
                                    key={`${depth}-${child.node!.pathID}`}
                                    {...props}
                                    node={child.node!}
                                    depth={depth + 1}
                                />
                            )
                        )}
                    </ul>
                )}
            </div>
        )
    },
    (previousProps, nextProps) => isEqual(previousProps, nextProps)
)
