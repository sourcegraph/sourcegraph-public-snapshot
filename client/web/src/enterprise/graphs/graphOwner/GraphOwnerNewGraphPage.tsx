import React, { useCallback, useMemo } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { NamespaceAreaContext } from '../../../namespaces/NamespaceArea'
import { NewGraphForm } from '../form/NewGraphForm'
import H from 'history'
import { GraphSelectionProps } from '../selector/graphSelectionProps'

interface Props extends NamespaceAreaContext, Pick<GraphSelectionProps, 'reloadGraphs'> {
    history: H.History
}

export const GraphOwnerNewGraphPage: React.FunctionComponent<Props> = ({ namespace, history, reloadGraphs }) => {
    const initialValue = useMemo<React.ComponentProps<typeof NewGraphForm>['initialValue']>(
        () => ({ owner: namespace.id, name: '', description: null, spec: '' }),
        [namespace.id]
    )
    const onCreate = useCallback((graph: Pick<GQL.IGraph, 'url'>) => history.push(graph.url), [history])
    return (
        <div className="container">
            <h2>New graph</h2>
            <NewGraphForm initialValue={initialValue} onCreate={onCreate} reloadGraphs={reloadGraphs} />
        </div>
    )
}
