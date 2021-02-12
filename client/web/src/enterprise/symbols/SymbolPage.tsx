import React from 'react'
import { Observable } from 'rxjs'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../backend/graphql'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { DocumentSymbolResult, DocumentSymbolVariables, SymbolPageSymbolFields } from '../../graphql-operations'
import { map } from 'rxjs/operators'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { useObservable } from '../../../../shared/src/util/useObservable'

const SymbolPageSymbolsGQLFragment = gql`
    fragment SymbolPageSymbolFields on DocSymbol {
        id
        text
        detail
        kind
        tags
        children {
            id
            text
            kind
            tags
        }
    }
`

const querySymbolUncached = (vars: DocumentSymbolVariables): Observable<SymbolPageSymbolFields[] | null> =>
    requestGraphQL<DocumentSymbolResult, DocumentSymbolVariables>(
        gql`
            query DocumentSymbol($repo: ID!, $commitID: String!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $commitID) {
                            tree(path: "") {
                                docSymbols {
                                    nodes {
                                        ...SymbolPageSymbolFields
                                    }
                                }
                            }
                        }
                    }
                }
            }
            ${SymbolPageSymbolsGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.commit?.tree?.docSymbols?.nodes || null)
    )

const querySymbol = memoizeObservable(querySymbolUncached, parameters => JSON.stringify(parameters))

export interface SymbolRouteProps extends Pick<RepoRevisionContainerContext, 'repo' | 'revision'> {}

export const SymbolPage: React.FunctionComponent<SymbolRouteProps> = ({ repo, revision }) => {
    const symbols = useObservable(querySymbol({ repo: repo.id, commitID: revision }))
    console.log('# symbols', symbols)
    return <div>Symbol page</div>
}
