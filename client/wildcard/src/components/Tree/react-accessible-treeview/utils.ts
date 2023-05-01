/* eslint-disable unicorn/no-abusive-eslint-disable */
/* eslint-disable */
import { useEffect, useRef } from 'react'

import { INode, INodeRef } from '.'

export type EventCallback = <T, E>(event: React.MouseEvent<T, E> | React.KeyboardEvent<T>) => void

export const composeHandlers =
    (...handlers: EventCallback[]): EventCallback =>
    (event): void => {
        for (const handler of handlers) {
            handler && handler(event)
            if (event.defaultPrevented) {
                break
            }
        }
    }

export const difference = (a: Set<number>, b: Set<number>) => {
    const s = new Set<number>()
    for (const v of a) {
        if (!b.has(v)) {
            s.add(v)
        }
    }
    return s
}

export const symmetricDifference = (a: Set<number>, b: Set<number>) => {
    return new Set<number>([...difference(a, b), ...difference(b, a)])
}

export const usePrevious = (x: Set<number>): Set<number> | undefined => {
    const ref = useRef<Set<number> | undefined>()
    useEffect(() => {
        ref.current = x
    }, [x])
    return ref.current
}

export const usePreviousData = (value: INode[] | undefined) => {
    const ref = useRef<INode[] | undefined>()
    useEffect(() => {
        ref.current = value
    })
    return ref.current
}

export const isBranchNode = (data: INode[], i: number) => data[i].children != null && data[i].children.length > 0

export const scrollToRef = (ref: INodeRef) => {
    if (ref != null && ref.scrollIntoView) {
        ref.scrollIntoView({ block: 'nearest' })
    }
}

export const focusRef = (ref: INodeRef) => {
    if (ref != null && ref.focus) {
        ref.focus({ preventScroll: true })
    }
}

export const getParent = (data: INode[], id: number) => {
    return data[id].parent
}

export const getDescendants = (data: INode[], id: number, disabledIds: Set<number>) => {
    const descendants: number[] = []
    const getDescendantsHelper = (data: INode[], id: number) => {
        const node = data[id]
        if (node.children == null) return
        for (const childId of node.children.filter(x => !disabledIds.has(x))) {
            descendants.push(childId)
            getDescendantsHelper(data, childId)
        }
    }
    getDescendantsHelper(data, id)
    return descendants
}

export const getSibling = (data: INode[], id: number, diff: number) => {
    const parentId = getParent(data, id)
    if (parentId != null) {
        const parent = data[parentId]
        const index = parent.children.indexOf(id)
        const siblingIndex = index + diff
        if (parent.children[siblingIndex]) {
            return parent.children[siblingIndex]
        }
    }
    return null
}

export const getLastAccessible = (data: INode[], id: number, expandedIds: Set<number>) => {
    let node = data[id]
    const isRoot = data[0].id === id
    if (isRoot) {
        node = data[data[id].children[data[id].children.length - 1]]
    }
    while (expandedIds.has(node.id) && isBranchNode(data, node.id)) {
        node = data[node.children[node.children.length - 1]]
    }
    return node.id
}

export const getPreviousAccessible = (data: INode[], id: number, expandedIds: Set<number>) => {
    if (id === data[0].children[0]) {
        return null
    }
    const previous = getSibling(data, id, -1)
    if (previous == null) {
        return getParent(data, id)
    }
    return getLastAccessible(data, previous, expandedIds)
}

export const getNextAccessible = (data: INode[], id: number, expandedIds: Set<number>) => {
    let nodeId: number | null = data[id].id
    if (isBranchNode(data, nodeId) && expandedIds.has(nodeId)) {
        return data[nodeId].children[0]
    }
    while (true) {
        const next = getSibling(data, nodeId, 1)
        if (next != null) {
            return next
        }
        nodeId = getParent(data, nodeId)

        //we have reached the root so there is no next accessible node
        if (nodeId == null) {
            return null
        }
    }
}

export const propagateSelectChange = (
    data: INode[],
    ids: Set<number>,
    selectedIds: Set<number>,
    disabledIds: Set<number>
) => {
    const changes = {
        every: new Set<number>(),
        some: new Set<number>(),
        none: new Set<number>(),
    }
    for (const id of ids) {
        let currentId = id
        while (true) {
            const parent = getParent(data, currentId)
            if (parent === 0 || parent == null || (parent != null && disabledIds.has(parent))) {
                break
            }
            const enabledChildren = data[parent].children.filter(x => !disabledIds.has(x))
            if (enabledChildren.length === 0) break
            const some = enabledChildren.some(x => selectedIds.has(x) || changes.some.has(x))
            if (!some) {
                changes.none.add(parent)
            } else {
                if (enabledChildren.every(x => selectedIds.has(x))) {
                    changes.every.add(parent)
                } else {
                    changes.some.add(parent)
                }
            }
            currentId = parent
        }
    }
    return changes
}

export const getAccessibleRange = ({
    data,
    expandedIds,
    from,
    to,
}: {
    data: INode[]
    expandedIds: Set<number>
    from: number
    to: number
}) => {
    const range: number[] = []
    const max_loop = Object.keys(data).length
    let count = 0
    let currentId: number | null = from
    range.push(from)
    if (from < to) {
        while (count < max_loop) {
            currentId = getNextAccessible(data, currentId, expandedIds)
            currentId != null && range.push(currentId)
            if (currentId == null || currentId === to) break
            count += 1
        }
    } else if (from > to) {
        while (count < max_loop) {
            currentId = getPreviousAccessible(data, currentId, expandedIds)
            currentId != null && range.push(currentId)
            if (currentId == null || currentId === to) break
            count += 1
        }
    }

    return range
}

interface ITreeNode {
    name: string
    children?: ITreeNode[]
}

export const flattenTree = function (tree: ITreeNode): INode[] {
    let count = 0
    const flattenedTree: INode[] = []

    const flattenTreeHelper = function (tree: ITreeNode, parent: number | null) {
        const node: INode = {
            id: count,
            name: tree.name,
            children: [],
            parent,
        }
        flattenedTree[count] = node
        count += 1
        if (tree.children == null || tree.children.length === 0) return
        for (const child of tree.children) {
            flattenTreeHelper(child, node.id)
        }
        node.children = flattenedTree.filter(x => x.parent === node.id).map((x: INode) => x.id)
    }

    flattenTreeHelper(tree, null)
    return flattenedTree
}

export const getAriaSelected = ({
    isSelected,
    isDisabled,
    multiSelect,
}: {
    isSelected: boolean
    isDisabled: boolean
    multiSelect: boolean
}): boolean | undefined => {
    if (isDisabled) return isSelected
    if (multiSelect) return isSelected
    return isSelected ? true : undefined
}

export const getAriaChecked = ({
    isSelected,
    isDisabled,
    isHalfSelected,
    multiSelect,
}: {
    isSelected: boolean
    isDisabled: boolean
    isHalfSelected: boolean
    multiSelect: boolean
}): boolean | undefined | 'mixed' => {
    if (isDisabled) return isSelected
    if (isHalfSelected) return 'mixed'
    if (multiSelect) return isSelected
    return isSelected ? true : undefined
}

export const propagatedIds = (data: INode[], ids: number[], disabledIds: Set<number>) =>
    ids.concat(...ids.filter(id => isBranchNode(data, id)).map(id => getDescendants(data, id, disabledIds)))

const isIE = () => window.navigator.userAgent.match(/Trident/)

export const onComponentBlur = (event: React.FocusEvent, treeNode: HTMLUListElement | null, callback: () => void) => {
    if (treeNode == null) {
        console.warn('ref not set on <ul>')
        return
    }
    if (isIE()) {
        setTimeout(() => !treeNode.contains(document.activeElement) && callback(), 0)
    } else {
        !treeNode.contains(event.nativeEvent.relatedTarget as Node) && callback()
    }
}
