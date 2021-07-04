import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../backend/graphql'
import { MonikersAtLocationVariables, MonikersAtLocationResult, Moniker } from '../../graphql-operations'

const MonikerGQLFragment = gql`
    fragment Moniker on Moniker {
        scheme
        identifier
    }
`

export const getMonikersAtLocation = (vars: MonikersAtLocationVariables): Observable<Moniker[] | null> =>
    requestGraphQL<MonikersAtLocationResult, MonikersAtLocationVariables>(
        gql`
            query MonikersAtLocation($repo: ID!, $commitID: String!, $path: String!, $line: Int!, $character: Int!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $commitID) {
                            blob(path: $path) {
                                lsif {
                                    monikers(line: $line, character: $character) {
                                        scheme
                                        identifier
                                    }
                                }
                            }
                        }
                    }
                }
            }
            ${MonikerGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.commit?.blob?.lsif?.monikers || null)
    )
