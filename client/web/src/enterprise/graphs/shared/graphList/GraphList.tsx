import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { GraphListItem, GraphListItemProps } from './GraphListItem'

interface Props extends GraphListItemProps {
    /** Graphs, or `undefined` if loading. */
    graphs: { nodes: React.ComponentProps<typeof GraphListItem>['node'][] } | undefined
}

export const GraphList: React.FunctionComponent<Props> = ({ graphs, ...props }) =>
    graphs ? (
        graphs.nodes.length > 0 ? (
            <ul className="list-group">
                {graphs.nodes.map(graph => (
                    <GraphListItem
                        {...props}
                        key={graph.id === null ? 'null' : graph.id}
                        node={graph}
                        className="list-group-item"
                    />
                ))}
            </ul>
        ) : (
            <div className="card">
                <p className="card-body mb-0 text-muted">No graphs.</p>
            </div>
        )
    ) : (
        <LoadingSpinner className="icon-inline" />
    )
