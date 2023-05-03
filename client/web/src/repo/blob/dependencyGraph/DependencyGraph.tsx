import { useCallback } from 'react'

import { forceSimulation, forceManyBody, forceCenter, forceLink } from 'd3-force'
import ReactFlow, {
    Controls,
    Background,
    Node,
    Edge,
    useNodesState,
    useEdgesState,
    addEdge,
    MarkerType,
} from 'reactflow'

import FloatingConnectionLine from './FloatingConnectionLine'
import FloatingEdge from './FloatingEdge'
import { getTreeData } from './utils'

import styles from './DependencyGraph.module.scss'

interface Props {
    repoID: string
    revision: string
    filePath: string
}

const edgeTypes = {
    floating: FloatingEdge,
}

function getLayoutedNodes(nodes: Node[], links: Edge[]): Node[] {
    // Calculate the degree for each node
    let nodeDegree: { [key: string]: number } = {}
    for (let link of links) {
        nodeDegree[link.source] = (nodeDegree[link.source] || 0) + 1
        nodeDegree[link.target] = (nodeDegree[link.target] || 0) + 1
    }

    // Find the node with the maximum degree
    let maxDegree = -1
    let centerNodeId: string | null = null
    for (let nodeId in nodeDegree) {
        if (nodeDegree[nodeId] > maxDegree) {
            maxDegree = nodeDegree[nodeId]
            centerNodeId = nodeId
        }
    }

    // Create a new simulation with a small charge to push nodes apart.
    const simulation = forceSimulation(nodes as any)
        .force('charge', forceManyBody().strength(-50))
        .force('center', forceCenter(200, 300)) // 1000px (width) / 2, 600px (height) / 2
        .force(
            'link',
            forceLink(links as any)
                .id((d: Node) => d.id)
                .distance(200)
        )
        .stop()

    // Find the node to be at the center and put it there
    let centerNode = nodes.find(n => n.id === centerNodeId)
    if (centerNode) {
        centerNode.fx = 200 // fx and fy set a fixed position for node
        centerNode.fy = 300
    }

    // Run the simulation for a sufficient number of iterations
    for (let i = 0; i < 300; ++i) simulation.tick()

    // After the simulation, copy the position values to the actual nodes
    nodes.forEach(node => {
        if (node.x && node.y) {
            node.position = { x: node.x, y: node.y }
        }
    })

    return nodes
}

export const DependencyGraph: React.FC<Props> = props => {
    const { nodes: initialNodes, links } = getTreeData(props.filePath)

    const [nodes, , onNodesChange] = useNodesState(getLayoutedNodes(initialNodes, JSON.parse(JSON.stringify(links))))
    const [edges, setEdges, onEdgesChange] = useEdgesState(links)

    const onConnect = useCallback(
        params => setEdges(eds => addEdge({ ...params, type: 'floating', markerEnd: { type: MarkerType.Arrow } }, eds)),
        [setEdges]
    )

    return (
        <div className={styles.container}>
            {/* <code>{JSON.stringify(data, null, 2)}</code> */}
            <div style={{ height: '100%' }}>
                <ReactFlow
                    fitView
                    nodes={nodes}
                    edges={edges}
                    edgeTypes={edgeTypes}
                    onNodesChange={onNodesChange}
                    onEdgesChange={onEdgesChange}
                    onConnect={onConnect}
                    connectionLineComponent={FloatingConnectionLine}
                >
                    <Background />
                    <Controls />
                </ReactFlow>
            </div>
        </div>
    )
}
