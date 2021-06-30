import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'

import { requestGraphQL } from '../../backend/graphql'
import { Scalars } from '../../graphql-operations'

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

export interface DocumentationPageResults {
    node: GQL.IRepository
}
export interface DocumentationPageVariables {
    repo: Scalars['ID']
    revspec: string
    pathID: string
}

export const fetchDocumentationPage = (args: DocumentationPageVariables): Observable<GQL.IDocumentationPage> =>
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
            const repo = data.node
            if (!repo.commit || !repo.commit.tree || !repo.commit.tree.lsif) {
                throw new Error('no LSIF data')
            }
            if (!repo.commit.tree.lsif.documentationPage || !repo.commit.tree.lsif.documentationPage.tree) {
                throw new Error('no LSIF documentation')
            }
            return repo.commit.tree.lsif.documentationPage
        })
    )
