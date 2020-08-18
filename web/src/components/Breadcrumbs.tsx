import React, { useState, useEffect, useMemo, useCallback } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Link } from '../../../shared/src/components/Link'

export interface Breadcrumb {
    key: string
    element: React.ReactNode | null
    divider?: React.ReactNode
}

export interface BreadcrumbsProps {
    breadcrumbs: BreadcrumbsByDepth
}

export interface UpdateBreadcrumbsProps {
    useSetBreadcrumb: UseSetBreadcrumb
    setBreadcrumb: SetBreadcrumb
}

type BreadcrumbOrFalsy = Breadcrumb | false | null | undefined

export type UseSetBreadcrumb = (breadcrumb: BreadcrumbOrFalsy) => BreadcrumbSetters

export type SetBreadcrumb = (breadcrumb: BreadcrumbOrFalsy) => BreadcrumbSetters & { cleanup: () => void }

interface BreadcrumbSetters {
    useSetBreadcrumb: UseSetBreadcrumb
    setBreadcrumb: SetBreadcrumb
}

type BreadcrumbsByDepth = [number, BreadcrumbOrFalsy][]

/**
 *
 *
 * TODO: Document how to use `useBreadcrumbs`
 */
export const useBreadcrumbs = (): {
    breadcrumbs: BreadcrumbsByDepth
    useSetBreadcrumb: UseSetBreadcrumb
    setBreadcrumb: SetBreadcrumb
} => {
    const [breadcrumbsByDepth, setBreadcrumbsByDepth] = useState<BreadcrumbsByDepth>([
        [0, { key: 'home', element: <Link to="/search">Home</Link>, divider: null }],
    ])

    const createBreadcrumbSetters = useCallback((depth: number = 1): BreadcrumbSetters => {
        /** Shared logic between plain function and hook */
        function _internalSetBreadcrumb(breadcrumb: BreadcrumbOrFalsy): () => void {
            const entry: [number, BreadcrumbOrFalsy] = [depth, breadcrumb]

            setBreadcrumbsByDepth(breadcrumbs => [...breadcrumbs, entry])
            // cleanup
            return () => {
                setBreadcrumbsByDepth(breadcrumbs => breadcrumbs.filter(breadcrumb => breadcrumb !== entry))
            }
        }

        /** Convenience hook for function components */
        function useSetBreadcrumb(breadcrumb: BreadcrumbOrFalsy): BreadcrumbSetters {
            useEffect(() => _internalSetBreadcrumb(breadcrumb), [breadcrumb])

            return useMemo(() => createBreadcrumbSetters(depth + 1), [])
        }

        /** 'Vanilla function' for backcompat with class components */
        function setBreadcrumb(breadcrumb: BreadcrumbOrFalsy): BreadcrumbSetters & { cleanup: () => void } {
            return {
                cleanup: _internalSetBreadcrumb(breadcrumb),
                ...createBreadcrumbSetters(depth + 1),
            }
        }

        return {
            useSetBreadcrumb,
            setBreadcrumb,
        }
    }, [])

    const breadcrumbSetters = useMemo(() => createBreadcrumbSetters(), [createBreadcrumbSetters])

    return {
        breadcrumbs: breadcrumbsByDepth,
        ...breadcrumbSetters,
    }
}

/** Renders breadcrumbs by depth */
export const Breadcrumbs: React.FC<{ breadcrumbs: BreadcrumbsByDepth }> = ({ breadcrumbs }) => {
    const nodes: React.ReactNode[] = []

    breadcrumbs.sort((a, b) => a[0] - b[0])

    for (const [_depth, breadcrumb] of breadcrumbs) {
        if (breadcrumb) {
            const divider =
                breadcrumb.divider === undefined ? <ChevronRightIcon className="icon-inline" /> : breadcrumb.divider
            nodes.push(
                <span key={breadcrumb.key} className="text-muted d-flex align-items-center">
                    <span className="font-weight-semibold">{divider}</span>
                    {breadcrumb.element}
                </span>
            )
        }
    }

    return (
        <nav className="d-flex" aria-label="Breadcrumbs">
            {nodes}
        </nav>
    )
}
