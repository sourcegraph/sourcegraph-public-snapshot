import * as H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'

import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'
import { RepositoryFields } from '../../graphql-operations'
import { toDocumentationURL } from '../../util/url'

import { DocumentationIcons } from './DocumentationIcons'
import { GQLDocumentationNode, isExcluded, Tag } from './graphql'

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
    const hash = hashIndex ? node.pathID.slice(hashIndex + '#'.length) : ''
    const thisPage = toDocumentationURL({ ...repoRevision, pathID: node.pathID })

    return (
        <div className="documentation-index-node">
            <DocumentationIndexNode
                key={`${depth}-content`}
                {...props}
                node={node}
                depth={depth + 1}
                contentOnly={true}
            />
        </div>
    )
}
