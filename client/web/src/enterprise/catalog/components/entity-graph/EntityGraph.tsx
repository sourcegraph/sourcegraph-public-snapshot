import classNames from 'classnames'
import { DagreReact, EdgeOptions, NodeOptions, RecursivePartial } from 'dagre-reactjs'
import React from 'react'

import { CatalogGraphFields } from '../../../../graphql-operations'

interface Props {
    graph: CatalogGraphFields
    className?: string
}

const defaultNodeConfig: RecursivePartial<NodeOptions> = {
    styles: {
        node: {
            padding: {
                top: 8,
                right: 16,
                bottom: 12,
                left: 16,
            },
        },
        label: {
            styles: { fill: 'var(--body-color)' },
        },
        shape: {
            styles: { strokeWidth: 0, fill: 'var(--merged-3)', fillOpacity: 1 },
        },
    },
}

const defaultEdgeConfig: RecursivePartial<EdgeOptions> = {
    styles: {
        label: {
            styles: { fill: 'var(--text-muted)' },
        },
        edge: {
            styles: { stroke: 'var(--border-color-2)', strokeWidth: '2.5px' },
        },
        marker: {
            styles: { fill: 'var(--border-color-2)' },
        },
    },
}

export const EntityGraph: React.FunctionComponent<Props> = ({ graph, className }) => (
    <svg width={1000} height={1000} className={classNames(className)}>
        <DagreReact
            nodes={graph.nodes.map(node => ({ id: node.id, label: node.name }))}
            edges={graph.edges.map(edge => ({ from: edge.outNode.id, to: edge.inNode.id, label: edge.outType }))}
            defaultNodeConfig={defaultNodeConfig}
            defaultEdgeConfig={defaultEdgeConfig}
        />
    </svg>
)
