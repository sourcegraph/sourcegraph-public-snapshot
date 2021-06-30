import * as H from 'history'
import React, { useMemo } from 'react'
import { Link } from 'react-router-dom'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'
import { RepositoryFields } from '../../graphql-operations'
import { toDocumentationURL } from '../../util/url'

import { DocumentationIcons } from './DocumentationIcons'

// Mirrors the same type on the backend:
//
// https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+type+DocumentationNode+struct&patternType=literal
export interface GQLDocumentationNode {
    pathID: string
    documentation: Documentation
    label: MarkupContent
    detail: MarkupContent
    children: DocumentationNodeChild[]
}

export interface MarkupContent {
    kind: MarkupKind
    value: string
}
export type MarkupKind = 'plaintext' | 'markdown'

export interface Documentation {
    identifier: string
    newPage: boolean
    searchKey: string
    tags: Tag[]
}

export type Tag =
    | 'private'
    | 'deprecated'
    | 'test'
    | 'benchmark'
    | 'example'
    | 'license'
    | 'owner'
    | 'file'
    | 'module'
    | 'namespace'
    | 'package'
    | 'class'
    | 'method'
    | 'property'
    | 'field'
    | 'constructor'
    | 'enum'
    | 'interface'
    | 'function'
    | 'variable'
    | 'constant'
    | 'string'
    | 'number'
    | 'boolean'
    | 'array'
    | 'object'
    | 'key'
    | 'null'
    | 'enumNumber'
    | 'struct'
    | 'event'
    | 'operator'
    | 'typeParameter'

export interface DocumentationNodeChild {
    node?: GQLDocumentationNode
    pathID?: string
}

export function isExcluded(node: GQLDocumentationNode, excludingTags: Tag[]): boolean {
    return node.documentation.tags.filter(tag => excludingTags.includes(tag)).length > 0
}

interface Props extends Partial<RevisionSpec>, ResolvedRevisionSpec, BreadcrumbSetters {
    repo: RepositoryFields

    history: H.History
    location: H.Location

    /** The documentation node to render */
    node: GQLDocumentationNode

    /** How far deep we are in the tree of documentation nodes */
    depth: number

    /** The pathID of the page containing this documentation node */
    pagePathID: string

    /** A list of documentation tags, a section will not be rendered if it matches one of these. */
    excludingTags: Tag[]
}

export const DocumentationNode: React.FunctionComponent<Props> = ({ useBreadcrumb, node, depth, ...props }) => {
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

    useBreadcrumb(
        useMemo(
            () =>
                depth === 0 ? { key: `node-${depth}`, element: <Link to={thisPage}>{node.label.value}</Link> } : null,
            [depth, node.label.value, thisPage]
        )
    )

    return (
        <div className="documentation-node">
            <Link className={`h${depth + 1 < 4 ? depth + 1 : 4}`} id={hash} to={thisPage}>
                <DocumentationIcons tags={node.documentation.tags} /> {node.label.value}
            </Link>
            {node.detail.value !== '' && (
                <div className="px-2 pt-2">
                    <Markdown dangerousInnerHTML={renderMarkdown(node.detail.value)} />
                </div>
            )}

            {node.children?.map(
                (child, index) =>
                    child.node &&
                    !isExcluded(child.node, props.excludingTags) && (
                        <DocumentationNode
                            key={`${depth}-${index}`}
                            {...props}
                            node={child.node}
                            depth={depth + 1}
                            useBreadcrumb={useBreadcrumb}
                        />
                    )
            )}
        </div>
    )
}
