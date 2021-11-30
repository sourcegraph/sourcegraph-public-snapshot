import classNames from 'classnames'
import { DagreReact, EdgeOptions, NodeOptions, RecursivePartial } from 'dagre-reactjs'
import React, { useRef, useState } from 'react'
import { UncontrolledReactSVGPanZoom } from 'react-svg-pan-zoom'
import AutoSizer from 'react-virtualized-auto-sizer'

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

export const EntityGraph: React.FunctionComponent<Props> = ({ graph, className }) => {
    const viewer = useRef<UncontrolledReactSVGPanZoom>(null)
    const [dimensions, setDimensions] = useState({ width: 1000, height: 1000 })
    return (
        <div className={classNames(className)} style={{ height: '500px' }}>
            <AutoSizer>
                {({ height, width }) => (
                    <UncontrolledReactSVGPanZoom
                        width={width}
                        height={height}
                        tool="pan"
                        background="transparent"
                        detectAutoPan={false}
                        miniatureProps={{
                            position: 'none',
                            background: '#fff',
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
                                nodes={graph.nodes.map(node => ({ id: node.id, label: node.name }))}
                                edges={graph.edges.map(edge => ({
                                    from: edge.outNode.id,
                                    to: edge.inNode.id,
                                    label: edge.outType,
                                }))}
                                defaultNodeConfig={defaultNodeConfig}
                                defaultEdgeConfig={defaultEdgeConfig}
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
                            />
                        </svg>
                    </UncontrolledReactSVGPanZoom>
                )}
            </AutoSizer>
        </div>
    )
}
