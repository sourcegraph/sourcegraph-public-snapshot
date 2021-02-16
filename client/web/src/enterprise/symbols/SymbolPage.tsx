import React, { useEffect } from 'react'
import { Observable } from 'rxjs'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../backend/graphql'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { map } from 'rxjs/operators'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { RouteComponentProps } from 'react-router'
import { SymbolsSidebarOptionsSetterProps } from './SymbolsArea'
import {
    DocSymbolFieldsFragment,
    DocumentSymbolResult,
    DocumentSymbolVariables,
    SymbolPageSymbolFields,
} from '../../graphql-operations'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../shared/src/util/markdown'
import { Link } from 'react-router-dom'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

export interface Symbol extends DocSymbolFieldsFragment {
    children?: Symbol[]
}

const SymbolPageSymbolsGQLFragment = gql`
    fragment DocSymbolFieldsFragment on DocSymbol {
        id
        text
        detail
        kind
        tags
        hover {
            markdown {
                text
            }
        }
        definitions {
            nodes {
                url
            }
        }
    }
    fragment DocSymbolHierarchyFragment on DocSymbol {
        ...DocSymbolFieldsFragment
        children {
            ...DocSymbolFieldsFragment
            children {
                ...DocSymbolFieldsFragment
                children {
                    ...DocSymbolFieldsFragment
                }
            }
        }
    }
    fragment SymbolPageSymbolFields on DocSymbol {
        ...DocSymbolHierarchyFragment
        root {
            ...DocSymbolHierarchyFragment
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

export interface SymbolRouteProps {
    symbolID: string
}

export interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'revision'>,
        SymbolsSidebarOptionsSetterProps,
        RouteComponentProps<SymbolRouteProps> {}

export const SymbolPage: React.FunctionComponent<Props> = ({
    repo,
    revision,
    match: {
        params: { symbolID },
    },
    history,
    setSidebarOptions,
}) => {
    const symbol = useObservable(querySymbol({ repo: repo.id, commitID: revision, symbolID }))
    useEffect(() => {
        setSidebarOptions({ containerSymbol: symbol?.root as Symbol })
        return () => setSidebarOptions(null)
    }, [symbol || null]) // TODO(beyang): may want to specify dependencies
    if (!symbol) {
        return <div>Symbol not found</div>
    }

    const hoverParts = symbol.hover?.markdown.text.split('---', 2)
    const hoverSig = hoverParts?.[0]
    const hoverDoc = hoverParts?.[1]

    return symbol === null ? (
        <p className="p-3 text-muted h3">Not found</p>
    ) : symbol === undefined ? (
        <LoadingSpinner className="m-3" />
    ) : (
        <>
            <div className="mx-3 mt-3">
                {hoverSig &&
                    (symbol.definitions.nodes.length > 0 ? (
                        <Link to={symbol.definitions.nodes[0].url}>
                            <Markdown
                                dangerousInnerHTML={renderMarkdown(hoverSig)}
                                history={history}
                                className={`symbol-hover__signature`}
                            />
                        </Link>
                    ) : (
                        <Markdown
                            dangerousInnerHTML={renderMarkdown(hoverSig)}
                            history={history}
                            className={`symbol-hover__signature`}
                        />
                    ))}
            </div>
            {hoverDoc && <Markdown dangerousInnerHTML={renderMarkdown(hoverDoc)} history={history} className="mx-3" />}
            {/* <div>Symbol: {symbol.text}</div>
            <div>Definition</div>
            <div>Detail: {symbol.detail}</div>
            <div>Examples</div>
            <div>Children</div> */}
        </>
    )
}
