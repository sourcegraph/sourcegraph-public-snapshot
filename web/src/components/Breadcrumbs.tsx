import React, { useMemo, useState } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

export interface BreadcrumbNode {
    readonly next: BreadcrumbNode | null
    readonly key: string
    readonly element: JSX.Element | null
}

export interface ParentBreadcrumbProps {
    parentBreadcrumb: UpdateableBreadcrumb
}

export interface RootBreadcrumbProps {
    rootBreadcrumb: BreadcrumbNode
}

export interface UpdateableBreadcrumb {
    breadcrumb: BreadcrumbNode
    setChildBreadcrumb: (key: string, element: JSX.Element) => UpdateableBreadcrumb
    removeChildBreadcrumb: () => void
}

function createUpdateableBreadcrumb(
    breadcrumb: BreadcrumbNode,
    setBreadcrumb: (breadcrumb: BreadcrumbNode) => void
): UpdateableBreadcrumb {
    return {
        breadcrumb,
        setChildBreadcrumb: (key, element) => {
            const next = { key, element, next: null }
            setBreadcrumb({ ...breadcrumb, next })
            return createUpdateableBreadcrumb(next, updatedChild =>
                setBreadcrumb({ ...breadcrumb, next: updatedChild })
            )
        },
        removeChildBreadcrumb: () => {
            setBreadcrumb({ ...breadcrumb, next: null })
        },
    }
}

export const useRootBreadcrumb = (): UpdateableBreadcrumb => {
    const [rootBreadcrumb, setRootBreadcrumb] = useState<BreadcrumbNode>(() => ({
        key: 'home',
        element: <>Home</>,
        next: null,
    }))
    return useMemo(() => createUpdateableBreadcrumb(rootBreadcrumb, setRootBreadcrumb), [rootBreadcrumb])
}

export const Breadcrumbs: React.FunctionComponent<{ root: BreadcrumbNode }> = ({ root }) => (
    <>
        {root.element}
        {mapBreadcrumbs(root.next, ({ element, key }) => (
            <React.Fragment key={key}>
                <ChevronRightIcon /> {element}
            </React.Fragment>
        ))}
    </>
)

function mapBreadcrumbs(node: BreadcrumbNode | null, iteratee: (node: BreadcrumbNode) => void): void {
    if (!node) {
        return
    }
    iteratee(node)
    mapBreadcrumbs(node.next, iteratee)
}
