import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useMemo } from 'react'
import { Switch, Route, RouteComponentProps, Redirect } from 'react-router'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { HeroPage } from '../../../components/HeroPage'
import { NamespaceAreaContext } from '../../../namespaces/NamespaceArea'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { GraphOwnerGraphResult, GraphOwnerGraphVariables } from '../../../graphql-operations'
import { GraphPage, GraphPageGQLFragment } from '../detail/GraphPage'
import { map } from 'rxjs/operators'
import { EditGraphPageGQLFragment, EditGraphPage } from '../edit/GraphOwnerEditGraphPage'
import { requestGraphQL } from '../../../backend/graphql'
import { GraphSelectionProps } from '../selector/graphSelectionProps'
import { GraphTitle, GraphTitleGQLFragment } from '../shared/GraphTitle'

interface Props
    extends RouteComponentProps<{ graphName: string }>,
        NamespaceAreaContext,
        Pick<GraphSelectionProps, 'reloadGraphs'> {
    history: H.History
    location: H.Location
}

export const GraphArea: React.FunctionComponent<Props> = ({
    history,
    location,
    match: {
        url: matchURL,
        params: { graphName },
    },
    namespace,
    reloadGraphs,
}) => {
    const graph = useObservable(
        useMemo(
            () =>
                requestGraphQL<GraphOwnerGraphResult, GraphOwnerGraphVariables>(
                    gql`
                        query GraphOwnerGraph($owner: ID!, $name: String!) {
                            node(id: $owner) {
                                ... on GraphOwner {
                                    graph(name: $name) {
                                        ...GraphTitle
                                        ...GraphPage
                                        ...EditGraphPage
                                    }
                                }
                            }
                        }
                        ${GraphTitleGQLFragment}
                        ${GraphPageGQLFragment}
                        ${EditGraphPageGQLFragment}
                    `,
                    { owner: namespace.id, name: graphName }
                ).pipe(
                    map(dataOrThrowErrors),
                    map(data => data.node?.graph)
                ),
            [graphName, namespace.id]
        )
    )
    return (
        <div>
            {graph === undefined ? (
                <LoadingSpinner />
            ) : graph === null ? (
                <HeroPage
                    icon={MapSearchIcon}
                    title="404: Not Found"
                    subtitle="Sorry, the requested graph was not found."
                />
            ) : (
                <>
                    <ErrorBoundary location={location}>
                        <div className="container">
                            <GraphTitle graph={graph} />
                        </div>
                        <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                            <Switch>
                                <Route path={matchURL} exact={true}>
                                    <GraphPage graph={graph} />
                                </Route>
                                <Route path={matchURL + '/edit'} exact={true}>
                                    <div className="container">
                                        <EditGraphPage
                                            graph={graph}
                                            onDeleteURL={`${namespace.url}/graphs`}
                                            history={history}
                                            reloadGraphs={reloadGraphs}
                                        />
                                    </div>
                                </Route>
                                <Route key="hardcoded-key">
                                    <Redirect to={matchURL} />
                                </Route>
                            </Switch>
                        </React.Suspense>
                    </ErrorBoundary>
                </>
            )}
        </div>
    )
}
