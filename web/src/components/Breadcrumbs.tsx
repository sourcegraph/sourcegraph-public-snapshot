import React, { useState, useEffect, useMemo, useCallback } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Link } from '../../../shared/src/components/Link'
import { sortBy } from 'lodash'
import { Unsubscribable } from 'sourcegraph'
import { isDefined } from '../../../shared/src/util/types'

export interface Breadcrumb {
    /** A unique key for the breadcrumb. */
    key: string

    /** The breadcrumb element being displayed. */
    element: React.ReactNode

    /**
     * Optionally a custom divider displayed before the element.
     * By default a chevron icon `>` is used.
     */
    divider?: React.ReactNode
}

/**
 * Props of the `Breadcrumbs` component.
 */
export interface BreadcrumbsProps {
    /**
     * All current breadcrumbs.
     */
    breadcrumbs: BreadcrumbAtDepth[]
}

type NullableBreadcrumb = Breadcrumb | null | undefined

/**
 * Exposes APIs for class and function components to register breadcrumbs.
 */
export interface BreadcrumbSetters {
    /**
     * Hook for function components to register a breadcrumb.
     *
     * @param breadcrumb The breadcrumb to register. If a falsy value is passed the breadcrumb will not be included.
     * @returns Another breadcrumb setters object to pass down to child components to register child breadcrumbs.
     */
    useBreadcrumbSetters: (breadcrumb: NullableBreadcrumb) => BreadcrumbSetters

    /**
     * Imperative method for class components to register a breadcrumb.
     *
     * @param breadcrumb The breadcrumb to register. If a falsy value is passed the breadcrumb will not be included.
     * @returns Another breadcrumb setters object to pass down to child components to register child breadcrumbs,
     * with a method to remove the breadcrumb again. The object should be added to a [subscription
     * bag](https://about.sourcegraph.com/handbook/engineering/languages/typescript#subscription-bag).
     */
    setBreadcrumb: (breadcrumb: NullableBreadcrumb) => BreadcrumbSetters & Unsubscribable
}

interface BreadcrumbAtDepth {
    /**
     * The position of the breadcrumb in the sequence of breadcrumbs
     */
    depth: number

    /**
     * The breadcrumb to render at the position.
     */
    breadcrumb: NullableBreadcrumb
}

/**
 * Hook to start a breadcrumb tree.
 *
 * TODO: Document how to use `useBreadcrumbs`
 */
export const useBreadcrumbs = (): BreadcrumbsProps & BreadcrumbSetters => {
    const [breadcrumbsByDepth, setBreadcrumbsByDepth] = useState<BreadcrumbAtDepth[]>([
        { depth: 0, breadcrumb: { key: 'home', element: <Link to="/search">Home</Link>, divider: null } },
    ])

    const createBreadcrumbSetters = useCallback((depth: number = 1): BreadcrumbSetters => {
        /** Shared logic between plain function and hook */
        function internalSetBreadcrumb(breadcrumb: NullableBreadcrumb): () => void {
            const entry: BreadcrumbAtDepth = { depth, breadcrumb }

            setBreadcrumbsByDepth(breadcrumbs => [...breadcrumbs, entry])
            // cleanup
            return () => {
                setBreadcrumbsByDepth(breadcrumbs => breadcrumbs.filter(breadcrumb => breadcrumb !== entry))
            }
        }

        /** Convenience hook for function components */
        function useBreadcrumbSetters(breadcrumb: NullableBreadcrumb): BreadcrumbSetters {
            useEffect(() => internalSetBreadcrumb(breadcrumb), [breadcrumb])

            return useMemo(() => createBreadcrumbSetters(depth + 1), [])
        }

        /** 'Vanilla function' for backcompat with class components */
        function setBreadcrumb(breadcrumb: NullableBreadcrumb): BreadcrumbSetters & Unsubscribable {
            const cleanup = internalSetBreadcrumb(breadcrumb)
            const setters = createBreadcrumbSetters(depth + 1)
            return { unsubscribe: cleanup, ...setters }
        }

        return {
            useBreadcrumbSetters,
            setBreadcrumb,
        }
    }, [])

    const breadcrumbSetters = useMemo(() => createBreadcrumbSetters(), [createBreadcrumbSetters])

    return {
        breadcrumbs: breadcrumbsByDepth,
        ...breadcrumbSetters,
    }
}

/**
 * Renders breadcrumbs by depth.
 */
export const Breadcrumbs: React.FC<{ breadcrumbs: BreadcrumbAtDepth[] }> = ({ breadcrumbs }) => (
    <nav className="d-flex" aria-label="Breadcrumbs">
        {sortBy(breadcrumbs, 'depth')
            .map(({ breadcrumb }) => breadcrumb)
            .filter(isDefined)
            .map(breadcrumb => {
                const divider =
                    breadcrumb.divider === undefined ? <ChevronRightIcon className="icon-inline" /> : breadcrumb.divider
                return (
                    <span key={breadcrumb.key} className="text-muted d-flex align-items-center">
                        <span className="font-weight-semibold">{divider}</span>
                        {breadcrumb.element}
                    </span>
                )
            })}
    </nav>
)
