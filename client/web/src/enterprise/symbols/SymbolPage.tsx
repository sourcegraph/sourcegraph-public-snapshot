import React from 'react'
import { Observable } from 'rxjs'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../backend/graphql'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { DocumentSymbolResult, DocumentSymbolVariables, SymbolPageSymbolFields } from '../../graphql-operations'
import { map } from 'rxjs/operators'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { RouteComponentProps } from 'react-router'

export interface SymbolRouteProps {
    symbolID: string
}

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

const querySymbolUncached = (vars: DocumentSymbolVariables): Observable<SymbolPageSymbolFields | null | undefined> =>
    requestGraphQL<DocumentSymbolResult, DocumentSymbolVariables>(
        gql`
            query DocumentSymbol($repo: ID!, $commitID: String!, $symbolID: String!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $commitID) {
                            tree(path: "") {
                                docSymbol(id: $symbolID) {
                                    ...SymbolPageSymbolFields
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
        map(data => data.node?.commit?.tree?.docSymbol)
    )

const querySymbol = memoizeObservable(querySymbolUncached, parameters => JSON.stringify(parameters))

export interface SymbolRouteProps
    extends Pick<RepoRevisionContainerContext, 'repo' | 'revision'>,
        RouteComponentProps<SymbolRouteProps> {
    symbolID: string
}

export const SymbolPage: React.FunctionComponent<SymbolRouteProps> = ({
    repo,
    revision,
    match: {
        params: { symbolID },
    },
}) => {
    const symbol = useObservable(querySymbol({ repo: repo.id, commitID: revision, symbolID }))
    if (!symbol) {
        return <div>Symbol not found</div>
    }
    console.log('# symbol', symbol)
    return (
        <>
            <div>Symbol: {symbol.text}</div>
            <div>Definition</div>
            <div>{symbol.detail}</div>
            <div>Examples</div>
            <div>Children</div>
        </>
    )
}
