import classNames from 'classnames'
import { CustomNodeLabelProps, DagreReact, EdgeOptions, NodeOptions, RecursivePartial } from 'dagre-reactjs'
import React, { useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { UncontrolledReactSVGPanZoom } from 'react-svg-pan-zoom'
import AutoSizer from 'react-virtualized-auto-sizer'

import { CatalogGraphFields } from '../../../../graphql-operations'
import { CatalogEntityIcon } from '../CatalogEntityIcon'

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

export const EntityGraph: React.FunctionComponent<Props> = ({ graph, className }) => {
    const viewer = useRef<UncontrolledReactSVGPanZoom>(null)
    const [dimensions, setDimensions] = useState({ width: 1000, height: 1000 })
    return (
        <div className={classNames(className)} style={{ height: '80vh' }}>
            <AutoSizer>
                {({ height, width }) => (
                    <UncontrolledReactSVGPanZoom
                        width={width}
                        height={height}
                        tool="auto"
                        background="transparent"
                        SVGBackground="transparent"
                        detectAutoPan={false}
                        miniatureProps={{
                            position: 'none',
                            background: 'transparent',
                            width: 100,
                            height: 100,
                        }}
                        toolbarProps={{
                            position: 'none',
                            SVGAlignX: undefined,
                            SVGAlignY: undefined,
                        }}
                        ref={viewer}
                    >
                        <svg width={dimensions.width} height={dimensions.height}>
                            <DagreReact
                                nodes={graph.nodes.map(node => ({
                                    id: node.id,
                                    label: node.name,
                                    labelType: 'Entity',
                                    meta: { entity: node },
                                }))}
                                edges={graph.edges.map(edge => ({
                                    from: edge.outNode.id,
                                    to: edge.inNode.id,
                                    label: edge.outType,
                                }))}
                                defaultNodeConfig={defaultNodeConfig}
                                defaultEdgeConfig={defaultEdgeConfig}
                                customNodeLabels={{
                                    Entity: {
                                        renderer: EntityNodeLabel,
                                        html: true,
                                    },
                                }}
                                graphLayoutComplete={(width?: number, height?: number) => {
                                    if (width && height) {
                                        setDimensions({ width, height })
                                    }
                                    setTimeout(() => {
                                        if (viewer.current) {
                                            viewer.current.fitToViewer()
                                        }
                                    })
                                }}
                                graphOptions={{
                                    marginx: 15,
                                    marginy: 15,
                                    rankdir: 'LR',
                                    ranksep: 55,
                                    nodesep: 15,
                                }}
                            />
                        </svg>
                    </UncontrolledReactSVGPanZoom>
                )}
            </AutoSizer>
        </div>
    )
}

const EntityNodeLabel: React.FunctionComponent<CustomNodeLabelProps> = ({
    node: {
        meta: { entity },
    },
}) => (
    <Link to={entity.url} className="d-flex align-items-center text-body text-nowrap">
        <CatalogEntityIcon entity={entity} className="icon-inline mr-1" /> {entity.name}
    </Link>
)
