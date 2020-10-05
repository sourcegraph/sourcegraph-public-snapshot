import React, { useCallback } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { gql } from '../../../../../shared/src/graphql/graphql'
import { EditGraphPage as EditGraphPageFragment } from '../../../graphql-operations'
import { EditGraphForm } from '../form/EditGraphForm'
import { GraphSelectionProps } from '../selector/graphSelectionProps'
import H from 'history'

export const EditGraphPageGQLFragment = gql`
    fragment EditGraphPage on Graph {
        id
        name
        description
        spec
    }
`

interface Props extends Pick<GraphSelectionProps, 'reloadGraphs'> {
    graph: EditGraphPageFragment

    /** The URL to navigate to after deletion. */
    onDeleteURL: string

    history: H.History
}

export const EditGraphPage: React.FunctionComponent<Props> = ({ graph, onDeleteURL, history, reloadGraphs }) => {
    const onUpdate = useCallback((graph: Pick<GQL.IGraph, 'url'>) => history.push(graph.url), [history])
    const onDelete = useCallback(() => history.push(onDeleteURL), [history, onDeleteURL])

    return (
        <div>
            <h2>Edit graph</h2>
            <EditGraphForm initialValue={graph} onUpdate={onUpdate} onDelete={onDelete} reloadGraphs={reloadGraphs} />
        </div>
    )
}
