import React, { useEffect } from 'react'
import { Observable, of } from 'rxjs'
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
import { FileLocations } from '../../../../branded/src/components/panel/views/FileLocations'
import { makeRepoURI } from '../../../../shared/src/util/url'
import { fetchHighlightedFileLineRanges } from '../../repo/backend'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { VersionContextProps } from '../../../../shared/src/search/util'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import { Location } from '@sourcegraph/extension-api-types'

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
        references {
            nodes {
                url
                range {
                    start {
                        line
                        character
                    }
                    end {
                        line
                        character
                    }
                }
                resource {
                    path
                    commit {
                        oid
                    }
                    repository {
                        name
                    }
                }
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
        SettingsCascadeProps,
        VersionContextProps,
        RouteComponentProps<SymbolRouteProps> {}

export const SymbolPage: React.FunctionComponent<Props> = ({
    repo,
    revision,
    match: {
        params: { symbolID },
    },
    history,
    location,
    settingsCascade,
    setSidebarOptions,
    ...props
}) => {
    const symbol = useObservable(querySymbol({ repo: repo.id, commitID: revision, symbolID }))
    useEffect(() => {
        setSidebarOptions({ containerSymbol: symbol?.root as Symbol })
        return () => setSidebarOptions(null)
    }, [symbol || null]) // TODO(beyang): may want to specify dependencies

    const hoverParts = symbol?.hover?.markdown.text.split('---', 2)
    const hoverSig = hoverParts?.[0]
    const hoverDoc = hoverParts?.[1]

    // NEXT
    console.log('# symbol', symbol, hoverSig, hoverDoc, symbol?.detail)

    return symbol === null ? (
        <p className="p-3 text-muted h3">Not found</p>
    ) : symbol === undefined ? (
        <LoadingSpinner className="m-3" />
    ) : (
        <>
            <div className="mx-3 mt-3">
                <style>{'pre code { font-size: 1.2rem; line-height: 2.5rem; }'}</style>
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
            <style>
                {
                    'td.line { display: none; } .code-excerpt .code { padding-left: 0.25rem !important; } .result-container__header { display: none; } .result-container { border: solid 1px var(--border-color) !important; border-width: 1px !important; margin: 1rem; }'
                }
            </style>
            {symbol.references.nodes.length > 1 && (
                <section id="refs" className="mt-2 mx-3">
                    <h2 className="mt-3 mb-2 h4">Examples</h2>
                    <FileLocations
                        location={location}
                        locations={of(
                            symbol.references.nodes.map<Location>(reference => ({
                                uri: makeRepoURI({
                                    repoName: reference.resource.repository.name,
                                    commitID: reference.resource.commit.oid,
                                    filePath: reference.resource.path,
                                }),
                                range: reference.range!,
                            }))
                        )}
                        icon={SourceRepositoryIcon}
                        isLightTheme={true} // TODO: pass through
                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                        settingsCascade={settingsCascade}
                        versionContext={props.versionContext}
                    />
                </section>
            )}
        </>
    )
}
