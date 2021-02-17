import React from 'react'
import { Link, NavLink } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { requestGraphQL } from '../../backend/graphql'
import { DocumentSymbolsVariables, DocSymbolFields, DocumentSymbolsResult } from '../../graphql-operations'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { ItemList, urlForSymbol } from './SymbolsSidebar'

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
            detail
            kind
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

    // return docSymbols ? <ItemList symbols={docSymbols} level={0} repo={repo} /> : <div>Symbols not found</div>
    return docSymbols ? (
        <div className="w-100">
            <h2 className="m-2">Packages</h2>
            <ul className="list-unstyled list-group list-group-flush w-100 border-top">
                {docSymbols.map(symbol => (
                    <li className="border-bottom">
                        <div className="py-1 px-2 d-flex" style={{ fontSize: '1rem' }}>
                            <SymbolIcon kind={symbol.kind} />
                            <div className="px-1">
                                <Link to={urlForSymbol(symbol, repo)}>
                                    <span className="text-truncate">{symbol.id}</span>
                                </Link>
                                <div>{symbol.detail.slice(0, firstSentenceLength(symbol.detail))}</div>
                            </div>
                        </div>
                    </li>
                ))}
            </ul>
        </div>
    ) : (
        <div>Symbols not found</div>
    )
    // return (
    //     <>
    //         {docSymbols?.map(symbol => (
    //             <div key={symbol.text}>
    //                 <NavLink to={urlForSymbol(symbol)}>{symbol.text}</NavLink>
    //                 <div>{symbol.detail}</div>
    //             </div>
    //         ))}
    //     </>
    // )
}

const isUpper = (string: string): boolean => /^\p{Lu}$/u.test(string)

// firstSentenceLen returns the length of the first sentence in s.
// The sentence ends after the first period followed by space and
// not preceded by exactly one uppercase letter.
//
// TODO(sqs): copied from go/doc package
/* eslint-disable id-length */
const firstSentenceLength = (s: string): number => {
    let ppp = ''
    let pp = ''
    let p = ''
    // eslint-disable-next-line @typescript-eslint/prefer-for-of
    for (let index = 0; index < s.length; index++) {
        let q = s[index]
        if (q === '\n' || q === '\r' || q === '\t') {
            q = ' '
        }
        if (q === ' ' && p === '.' && (!isUpper(pp) || isUpper(ppp))) {
            return index
        }
        if (p === '。' || p === '．') {
            return index
        }
        ppp = pp
        pp = p
        p = q
    }
    return s.length
    /* eslint-enable id-length */
}
