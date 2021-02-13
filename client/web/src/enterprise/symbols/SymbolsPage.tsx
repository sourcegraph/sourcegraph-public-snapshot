import React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { requestGraphQL } from '../../backend/graphql'
import { DocumentSymbolsVariables, DocSymbolFields, DocumentSymbolsResult } from '../../graphql-operations'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'

// export const SymbolsPage: React.FunctionComponent<Props> = ({ repo, resolvedRev, viewOptions, history, ...props }) => {
//     useEffect(() => {
//         eventLogger.logViewEvent('Symbols')
//     }, [])
//
//     const data = useObservable(
//         useMemo(
//             () =>
//                 queryRepositorySymbols({
//                     repo: repo.id,
//                     commitID: resolvedRev.commitID,
//                     path: '.',
//                     filters: viewOptions,
//                 }),
//             [repo.id, resolvedRev.commitID, viewOptions]
//         )
//     )
//
//     return data ? <ContainerSymbolsList symbols={data} history={history} /> : <LoadingSpinner className="m-3" />
// }

const SymbolsPageSymbolsGQLFragment = gql`
    fragment DocSymbolFields on DocSymbol {
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

const queryRepositorySymbolsUncached = (vars: DocumentSymbolsVariables): Observable<DocSymbolFields[] | null> =>
    requestGraphQL<DocumentSymbolsResult, DocumentSymbolsVariables>(
        gql`
            query DocumentSymbols($repo: ID!, $commitID: String!, $path: String!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $commitID) {
                            tree(path: $path) {
                                docSymbols {
                                    nodes {
                                        ...DocSymbolFields
                                    }
                                }
                            }
                        }
                    }
                }
            }
            ${SymbolsPageSymbolsGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.commit?.tree?.docSymbols?.nodes || null)
    )

const queryRepositorySymbols = memoizeObservable(queryRepositorySymbolsUncached, parameters =>
    JSON.stringify(parameters)
)

export interface SymbolsRouteProps extends Pick<RepoRevisionContainerContext, 'repo' | 'revision' | 'resolvedRev'> {}

export const SymbolsPage: React.FunctionComponent<SymbolsRouteProps> = ({ repo, revision }) => {
    const docSymbols = useObservable(queryRepositorySymbols({ repo: repo.id, commitID: revision, path: '' }))

    function urlForSymbol(symbol: DocSymbolFields): string {
        // TODO(beyang): this is a hack
        return `/${repo.name}/-/docs/${symbol.id}`
    }
    return (
        <>
            {docSymbols?.map(symbol => (
                <div key={symbol.text}>
                    <div>
                        <a href={urlForSymbol(symbol)}>{symbol.text}</a>
                    </div>
                    <div>{symbol.detail}</div>
                </div>
            ))}
        </>
    )
}
