import classNames from 'classnames'
import * as d3 from 'd3-shape'
import { CustomNodeLabelProps, DagreReact, EdgeOptions, NodeOptions, Point, RecursivePartial } from 'dagre-reactjs'
import React, { useEffect, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { UncontrolledReactSVGPanZoom } from 'react-svg-pan-zoom'
import AutoSizer from 'react-virtualized-auto-sizer'

import { CatalogGraphFields } from '../../../../graphql-operations'
import { catalogRelationTypeDisplayName } from '../../core/edges'
import { ComponentIcon } from '../ComponentIcon'

interface Props {
    graph: CatalogGraphFields
    activeNodeID?: string
    className?: string
}

const defaultNodeConfig: RecursivePartial<NodeOptions> = {
    styles: {
        node: {
            padding: {
                top: 8,
                right: 16,
                bottom: 8,
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
    labelOffset: 4,
    pathType: 'd3curve',
    styles: {
        label: {
            styles: { fill: 'var(--text-muted)', fontSize: '0.6rem' },
        },
        edge: {
            styles: { stroke: 'var(--text-muted)', strokeWidth: '2.5px', fill: 'transparent' },
        },
        marker: {
            styles: { fill: 'var(--text-muted)' },
        },
    },
}

export const EntityGraph: React.FunctionComponent<Props> = ({ graph, activeNodeID, className }) => {
    const [stage, setStage] = useState(0)
    useEffect(() => setStage(stage => stage + 1), [graph])

    const nodes: RecursivePartial<NodeOptions>[] = useMemo(
        () =>
            graph.nodes.map(node => ({
                id: node.id,
                label: node.name,
                labelType: 'Entity',
                meta: { entity: node, isActive: activeNodeID === node.id },
                styles:
                    activeNodeID === node.id
                        ? {
                              shape: {
                                  styles: {
                                      fill: 'var(--primary)',
                                      stroke: 'var(--body-color)',
                                      strokeWidth: '1.5px',
                                  },
                              },
                          }
                        : activeNodeID
                        ? { shape: { styles: { fillOpacity: 0.35 } } }
                        : undefined,
            })),
        [activeNodeID, graph.nodes]
    )
    const edges: RecursivePartial<EdgeOptions>[] = useMemo(() => {
        const edges = new Map<string, RecursivePartial<EdgeOptions>>()
        for (const edge of graph.edges) {
            const key = `${edge.outNode.id}-${edge.inNode.id}`
            const existing = edges.get(key)
            if (existing) {
                existing.label += `, ${edge.type}`
            } else {
                edges.set(key, {
                    label: catalogRelationTypeDisplayName(edge.type),
                    from: edge.outNode.id,
                    to: edge.inNode.id,
                    styles:
                        activeNodeID && edge.inNode.id !== activeNodeID && edge.outNode.id !== activeNodeID
                            ? {
                                  edge: { styles: { strokeOpacity: 0.5, strokeWidth: '1px' } },
                                  marker: { styles: { fillOpacity: 0.5 } },
                              }
                            : undefined,
                })
            }
        }
        return [...edges.values()]
    }, [activeNodeID, graph.edges])

    const viewer = useRef<UncontrolledReactSVGPanZoom>(null)
    const [dimensions, setDimensions] = useState({ width: 500, height: 500 })
    return (
        <div
            className={classNames(className)}
            style={{ height: `${dimensions.height}px`, visibility: dimensions.width === 500 ? 'hidden' : '' }}
        >
            <AutoSizer>
                {({ height, width }) => (
                    <UncontrolledReactSVGPanZoom
                        width={width}
                        height={height}
                        tool="auto"
                        background="transparent"
                        SVGBackground="transparent"
                        detectAutoPan={false}
                        scaleFactorMax={1}
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
                                stage={stage}
                                nodes={nodes}
                                edges={edges}
                                defaultNodeConfig={defaultNodeConfig}
                                defaultEdgeConfig={defaultEdgeConfig}
                                customNodeLabels={{
                                    Entity: {
                                        renderer: EntityNodeLabel,
                                        html: true,
                                    },
                                }}
                                customPathGenerators={{
                                    d3curve: generatePathD3Curve,
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
                                    marginx: 32,
                                    marginy: 32,
                                    rankdir: 'LR',
                                    ranksep: 75,
                                    nodesep: 25,
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
        meta: { entity, isActive },
    },
}) => (
    <Link
        to={entity.url}
        className={classNames('d-flex align-items-center text-body text-nowrap', { 'font-weight-bold': isActive })}
    >
        <ComponentIcon entity={entity} className="icon-inline mr-1" /> {entity.name}
    </Link>
)

const generatePathD3Curve = (points: Point[]): string => {
    const coords: [number, number][] = points.map(point => [point.x, point.y])
    const curve = d3.line().curve(d3.curveBasis)(coords)!
    return curve
}
