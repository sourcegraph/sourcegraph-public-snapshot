import * as H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'

import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'
import { RepositoryFields } from '../../graphql-operations'
import { toDocumentationURL } from '../../util/url'

import { DocumentationIcons } from './DocumentationIcons'
import { GQLDocumentationNode, isExcluded, Tag } from './DocumentationNode'

interface Props extends Partial<RevisionSpec>, ResolvedRevisionSpec {
    repo: RepositoryFields

    history: H.History
    location: H.Location

    /** The documentation node to render */
    node: GQLDocumentationNode

    /** How far deep we are in the tree of documentation nodes */
    depth: number

    /** The pathID of the page containing this documentation node */
    pagePathID: string

    /** If true, render subpage index only */
    subpagesOnly: boolean

    /** If true, render content index only */
    contentOnly: boolean

    /** A list of documentation tags, a section will not be rendered if it matches one of these. */
    excludingTags: Tag[]
}

export const DocumentationIndexNode: React.FunctionComponent<Props> = ({ node, depth, ...props }) => {
    useScrollToLocationHash(props.location)
    const repoRevision = {
        repoName: props.repo.name,
        revision: props.revision || '',
    }
    const hashIndex = node.pathID.indexOf('#')
    const hash = hashIndex !== -1 ? node.pathID.slice(hashIndex + '#'.length) : ''
    let path = hashIndex !== -1 ? node.pathID.slice(0, hashIndex) : node.pathID
    path = path === '/' ? '' : path
    const thisPage = toDocumentationURL({ ...repoRevision, pathID: path + '#' + hash })
    const excluded = isExcluded(node, props.excludingTags)
    if (excluded) {
        return null
    }
    if (props.subpagesOnly) {
        return (
            <div className="documentation-index-node">
                <ul className="pl-3">
                    {node.children?.map((child, index) =>
                        child.pathID ? (
                            <div key={`${depth}-${index}`} className="text-nowrap">
                                <Link to={toDocumentationURL({ ...repoRevision, pathID: child.pathID })}>
                                    {child.pathID.slice('/'.length) + '/'}
                                </Link>
                            </div>
                        ) : null
                    )}
                </ul>
            </div>
        )
    }
    if (props.contentOnly) {
        return (
            <div className="documentation-index-node">
                <Link id={'index-' + hash} to={thisPage} className="text-nowrap">
                    <DocumentationIcons tags={node.documentation.tags} /> {node.label.value}
                </Link>
                <ul className="pl-3">
                    {node.children?.map((child, index) =>
                        child.pathID ? null : (
                            <DocumentationIndexNode
                                key={`${depth}-${index}`}
                                {...props}
                                node={child.node!}
                                depth={depth + 1}
                                subpagesOnly={false}
                                contentOnly={true}
                            />
                        )
                    )}
                </ul>
            </div>
        )
    }

    return (
        <div className="documentation-index-node">
            <Link id="index-subpages" to={thisPage} className="text-nowrap">
                Subpages
            </Link>
            <DocumentationIndexNode
                key={`${depth}-subpages`}
                {...props}
                node={node}
                depth={depth + 1}
                subpagesOnly={true}
                contentOnly={false}
            />
            <DocumentationIndexNode
                key={`${depth}-content`}
                {...props}
                node={node}
                depth={depth + 1}
                subpagesOnly={false}
                contentOnly={true}
            />
        </div>
    )
}
