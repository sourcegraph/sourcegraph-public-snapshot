// SCIP index as object for https://sourcegraph.com/github.com/codecov/sourcegraph-codecov@92d2f701f935b7ce3c3504ab893f808643e6eb24/-/blob/src/insights.ts
import { uniqBy } from 'lodash'
import { MarkerType } from 'reactflow'

import scipIndex from './index.json'
import { DescriptorSuffix, parseSymbol, SCIPDocument, SCIPOccurrence, SCIPSymbol } from './SymbolParser'

// import { parseSymbol as parseSymbol2 } from './SymbolParser2'

/**
 * Returns scip document as object, see [Document.toObject](https://sourcegraph.com/github.com/sourcegraph/scip@d62dfc4d962f4ac975429e0fbb0ebdda25b46503/-/blob/bindings/typescript/scip.ts?L614-634).
 */
export function getDocument(path: string): SCIPDocument | undefined {
    // TODO: replace with API call
    return scipIndex.documents.find(d => d.relative_path === path) as SCIPDocument | undefined
}

export function getTreeData(path: string): any {
    const document = getDocument(path)

    if (!document) {
        return null
    }

    return getNodesAndEdges(
        uniqBy(document.occurrences, o => o.symbol),
        path
    )

    return buildDependencyTreeData(
        uniqBy(document.occurrences, o => o.symbol),
        path
    )
}

type Result = { package: string; module: string; symbols: string[] }

type ModuleTree = {
    id: string
    symbol: SCIPSymbol
    children: ModuleTree[]
}

interface GraphNode {
    id: string
    children: GraphNode[]
}

interface Edge {
    id: string
    source: string
    target: string
}

function getNodesAndEdges(occurrences: SCIPOccurrence[], path: string): { nodes: GraphNode[]; edges: Edge[] } {
    const map = new Map<string, Result>()
    const nodes = []
    const links = []

    for (const { symbol } of occurrences) {
        console.log(symbol)
        const parsedSymbol = parseSymbol(symbol)

        const fileName = parsedSymbol
            .descriptors!.filter(d => d.suffix === DescriptorSuffix.Namespace)
            .map(d => d.name)
            .join('/')

        let symbolNameParts = []
        const descriptors = parsedSymbol.descriptors!.filter(d =>
            [DescriptorSuffix.Namespace, DescriptorSuffix.Meta, DescriptorSuffix.Local, DescriptorSuffix.Macro].every(
                s => d.suffix !== s
            )
        )
        for (let i = 0; i < descriptors.length; i++) {
            if (descriptors[i].name) {
                symbolNameParts.push(descriptors[i].name)
            }
        }
        if (symbolNameParts.length === 0) continue

        const key = `${parsedSymbol.package?.name}/${fileName}`
        let item = map.get(key)
        if (!item) {
            item = { package: parsedSymbol.package?.name || '', module: fileName, symbols: [] }
            map.set(key, item)
            nodes.push({ id: fileName, data: { label: key }, position: { x: 0, y: 0 } })
            if (fileName !== path) {
                links.push({
                    id: fileName + path,
                    source: fileName,
                    target: path,
                    type: 'floating',
                    markerEnd: {
                        type: MarkerType.Arrow,
                    },
                })
            }
        }
        item.symbols.push(symbolNameParts.join('.'))
    }

    return { nodes, links }
}

type NestedObject = {
    [key: string]: NestedObject
}

function buildDependencyTreeData(occurrences: SCIPDocument['occurrences'], path: string) {
    const result: NestedObject = {}

    for (const { symbol } of occurrences) {
        const parsedSymbol = parseSymbol(symbol)

        const topLevelKey = parsedSymbol
            .descriptors!.filter(d => d.suffix === DescriptorSuffix.Namespace)
            .map(d => d.name)
            .join('/')

        if (!topLevelKey || topLevelKey === path) continue

        if (!result[topLevelKey]) {
            result[topLevelKey] = {}
        }

        let currentLevel = result[topLevelKey]

        const descriptors = parsedSymbol.descriptors!.filter(d =>
            [DescriptorSuffix.Namespace, DescriptorSuffix.Meta, DescriptorSuffix.Local, DescriptorSuffix.Macro].every(
                s => d.suffix !== s
            )
        )
        for (let i = 0; i < descriptors.length; i++) {
            const descriptor = descriptors[i]
            if (i < descriptors.length - 1) {
                const key = descriptor.name
                if (!key) continue
                if (!currentLevel[key]) {
                    currentLevel[key] = {}
                }
                currentLevel = currentLevel[key]
            }
        }
    }

    return result
}
