import * as H from 'history'
import React, { useMemo } from 'react'

import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepositoryFields } from '../../graphql-operations'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Link } from 'react-router-dom'
import { toDocsURL } from '../../util/url'
import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'
import LockIcon from 'mdi-react/LockIcon'
import CancelIcon from 'mdi-react/CancelIcon'

// Mirrors the same type on the backend:
//
// https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+type+DocumentationNode+struct&patternType=literal
export type GQLDocumentationNode = {
    pathID: string,
    documentation: Documentation,
    label: MarkupContent,
    detail: MarkupContent,
    children: DocumentationNodeChild[],
};

export type MarkupContent = {
    kind: MarkupKind,
    value: string,
};
export type MarkupKind = 'plaintext' | 'markdown' ;

export type Documentation = {
    slug: string,
    newPage: boolean,
    tags: DocumentationTag[],
};

export type DocumentationTag = 'exported' | 'unexported' | 'deprecated' ;

export type DocumentationNodeChild = {
    node?: GQLDocumentationNode,
    pathID?: string,
}

interface Props
    extends Partial<RevisionSpec>,
        ResolvedRevisionSpec,
        BreadcrumbSetters {
    repo: RepositoryFields

    history: H.History
    location: H.Location
    node: GQLDocumentationNode,
    depth: number,
    pagePathID: string
}

export const DocumentationNode: React.FunctionComponent<Props> = ({ useBreadcrumb, node, depth, ...props }) => {
    useScrollToLocationHash(props.location)
    const repoRev = {
        repoName: props.repo.name,
        revision: props.revision || '',
    }
    const hash = node.pathID.slice(props.pagePathID.length + '/'.length).replace(/\//g, '-')
    const thisPage = toDocsURL({...repoRev, pathID: props.pagePathID}) + (hash ? '#' + hash : '')
    if (depth == 0) {
        console.log(depth, node.pathID)
        useBreadcrumb(useMemo(() => ({ key: `node-${depth}`, element: <Link to={thisPage}>{node.label.value}</Link> }), []))
    }

    const tagIcons = {
        'exported': null,
        'unexported': <LockIcon className="icon-inline" data-tooltip="Unexported" />,
        'deprecated': <CancelIcon className="icon-inline" data-tooltip="Deprecated" />,
    }
    return (
        <div className="documentation-node">
            <Link
                className={`h${depth+1 < 4 ? depth+1 : 4}`}
                id={hash}
                to={thisPage}
            >
                {node.label.value}
                {node.documentation.tags && node.documentation.tags.map(tag => tagIcons[tag])}
            </Link>
            {node.detail.value !== '' && <div className="px-2 pt-2">
                <Markdown dangerousInnerHTML={renderMarkdown(node.detail.value)} />
            </div>}

            {node.children && node.children.map((child, i) =>
                child.node && <DocumentationNode
                    key={`${depth}-${i}`} 
                    {...props}
                    node={child.node!}
                    depth={depth+1}
                    useBreadcrumb={useBreadcrumb}
                />
            )}
        </div>
    )
}
