import gql from 'tagged-template-noop'

import type * as sourcegraph from '../api'
import { cache } from '../util'
import { type QueryGraphQLFn, queryGraphQL as sgQueryGraphQL } from '../util/graphql'

import { type GenericLSIFResponse, queryLSIF } from './api'

export const stencil = async (
    uri: string,
    hasStencilSupport: () => Promise<boolean>,
    queryGraphQL: QueryGraphQLFn<GenericLSIFResponse<{ stencil: sourcegraph.Range[] }>> = sgQueryGraphQL
): Promise<sourcegraph.Range[] | undefined> => {
    if (!(await hasStencilSupport())) {
        return undefined
    }

    const response = await queryLSIF({ query: stencilQuery, uri }, queryGraphQL)
    return response?.stencil
}

const stencilQuery = gql`
    query LegacyStencil($repository: String!, $commit: String!, $path: String!) {
        repository(name: $repository) {
            commit(rev: $commit) {
                blob(path: $path) {
                    lsif {
                        stencil {
                            start {
                                line
                                character
                            }
                            end {
                                line
                                character
                            }
                        }
                    }
                }
            }
        }
    }
`

export type StencilFn = (uri: string) => Promise<sourcegraph.Range[] | undefined>

export const makeStencilFn = (
    queryGraphQL: QueryGraphQLFn<GenericLSIFResponse<{ stencil: sourcegraph.Range[] }>>,
    hasStencilSupport: () => Promise<boolean> = () => Promise.resolve(true)
): StencilFn => cache(uri => stencil(uri, hasStencilSupport, queryGraphQL), { max: 10 })
