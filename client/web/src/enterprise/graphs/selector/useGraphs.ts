import { map } from 'rxjs/operators'
import { useMemo } from 'react'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { ViewerGraphsResult, ViewerGraphsVariables } from '../../../graphql-operations'
import { GraphSelectionProps, SelectableGraph } from './graphSelectionProps'
import { requestGraphQL } from '../../../backend/graphql'

/**
 * A React hook that returns the list of graphs to display for the viewer to select.
 */
export const useGraphs = ({
    reloadGraphsSeq,
    contextualGraphs,
}: Pick<GraphSelectionProps, 'reloadGraphsSeq' | 'contextualGraphs'>): SelectableGraph[] => {
    const viewerGraphs = useObservable(
        useMemo(
            () =>
                requestGraphQL<ViewerGraphsResult, ViewerGraphsVariables>(
                    gql`
                        query ViewerGraphs {
                            graphs(affiliated: true) {
                                nodes {
                                    id
                                    name
                                    description
                                    url
                                    editURL
                                }
                            }
                        }
                    `,
                    {}
                ).pipe(
                    map(dataOrThrowErrors),
                    map(data => data.graphs.nodes)
                ),
            [reloadGraphsSeq] // reload when GraphSelectorProps#reloadGraph is called
        )
    )

    const allGraphs = useMemo<SelectableGraph[]>(
        () =>
            window.context?.graphsEnabled
                ? [
                      { id: null, name: 'Everything', description: null },
                      ...(contextualGraphs || []),
                      ...(viewerGraphs || []),
                  ]
                : [],
        [contextualGraphs, viewerGraphs]
    )
    return allGraphs
}
