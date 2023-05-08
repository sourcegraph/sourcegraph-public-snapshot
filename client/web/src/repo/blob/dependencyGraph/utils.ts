// SCIP index as object for https://sourcegraph.com/github.com/codecov/sourcegraph-codecov@92d2f701f935b7ce3c3504ab893f808643e6eb24/-/blob/src/insights.ts
import { uniqBy } from 'lodash'

import scipIndex from './index.json'
import { DescriptorSuffix, parseSymbol, SCIPDocument, SCIPOccurrence, SCIPSymbol } from './SymbolParser'

interface Edge {
    id: string
    source: string
    target: string
}

interface GraphNode {
    name: string
    suffix: DescriptorSuffix
    symbol?: string
    children: Graph
}

type Graph = Map<string, GraphNode>

/**
 * Returns scip document as object, see [Document.toObject](https://sourcegraph.com/github.com/sourcegraph/scip@d62dfc4d962f4ac975429e0fbb0ebdda25b46503/-/blob/bindings/typescript/scip.ts?L614-634).
 */
export function getDocument(path: string): SCIPDocument | undefined {
    // TODO: replace with API call
    return scipIndex.documents.find(d => d.relative_path === path) as SCIPDocument | undefined
}

export function getTreeData(path: string): ReturnType<typeof getNodesAndEdges> | null {
    const document = getDocument(path)

    if (!document) {
        return null
    }

    return getNodesAndEdges(
        uniqBy(document.occurrences, o => o.symbol),
        path
    )
}

function createGraph(data: [string, SCIPSymbol][]): Graph {
    const graph: Graph = new Map()

    for (const [symbol, { descriptors }] of data) {
        let fileName = ''
        let currentNode: GraphNode | undefined = undefined

        for (let i = 0; i < descriptors!.length; i++) {
            let node: GraphNode | undefined = undefined
            const descriptor = descriptors![i]
            switch (descriptor.suffix) {
                case DescriptorSuffix.Namespace: {
                    const nextDescriptor = descriptors![i + 1]
                    fileName += fileName ? '/' + descriptor.name : descriptor.name
                    if (!nextDescriptor || nextDescriptor.suffix !== DescriptorSuffix.Namespace) {
                        node = graph.get(fileName)
                        if (!node) {
                            node = { name: fileName, suffix: DescriptorSuffix.Namespace, children: new Map() }
                            graph.set(fileName, node)
                        }
                        currentNode = node
                    }
                    break
                }

                case DescriptorSuffix.Meta:
                case DescriptorSuffix.Local:
                case DescriptorSuffix.Macro:
                    break

                default: {
                    node = currentNode!.children.get(descriptor.name)
                    if (!node) {
                        node = { name: descriptor.name, suffix: descriptor.suffix, children: new Map() }
                        currentNode!.children.set(descriptor.name, node)
                    }
                    currentNode = node
                    break
                }
            }

            if (node && i === descriptors!.length - 1) {
                node.symbol = symbol
            }
        }
    }

    return graph
}

function getNodesAndEdges(
    occurrences: SCIPOccurrence[],
    path: string
): { nodes: { id: string; data: GraphNode }[]; edges: Edge[] } {
    const graph = createGraph(occurrences.map(o => [o.symbol, parseSymbol(o.symbol)]))
    const nodes = []
    const edges = []

    for (const [fileName, data] of graph) {
        nodes.push({ id: fileName, data })

        if (fileName !== path) {
            edges.push({ id: fileName, source: fileName, target: path })
        }
    }

    return { nodes, edges }
}
