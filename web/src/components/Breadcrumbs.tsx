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
     * @param breadcrumb The breadcrumb to register. If a falsy value is passed the breadcrumb will not be included. **NOTE: The argument MUST be wrapped in `useMemo()`**.
     * @returns Another breadcrumb setters object to pass down to child components to register child breadcrumbs.
     */
    useBreadcrumb: (breadcrumb: NullableBreadcrumb) => BreadcrumbSetters

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
 * @returns An object with `breadcrumbs` to be passed to the `Breadcrumbs` component and a pair of breadcrumb setters
 * to pass down to child components to register child breadcrumbs. Be sure to pass down the returned breadcrumb setters,
 * not the setters that were passed to the component. Otherwise, your breadcrumbs may render out of order.
 *
 */
export const useBreadcrumbs = (): BreadcrumbsProps & BreadcrumbSetters => {
    const [breadcrumbsByDepth, setBreadcrumbsByDepth] = useState<BreadcrumbAtDepth[]>([
        { depth: 0, breadcrumb: { key: 'home', element: <Link to="/search">Home</Link>, divider: null } },
    ])

    /**
     * @param depth The relative depth of the next breadcrumb to be added with the
     * returned breadcrumb setters. This should always be called with $CURRENT_DEPTH + 1.
     */
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
        function useBreadcrumb(breadcrumb: NullableBreadcrumb): BreadcrumbSetters {
            useEffect(() => internalSetBreadcrumb(breadcrumb), [breadcrumb])

            return useMemo(() => createBreadcrumbSetters(depth + 1), [])
        }

        /** Plain function for backcompat with class components */
        function setBreadcrumb(breadcrumb: NullableBreadcrumb): BreadcrumbSetters & Unsubscribable {
            const cleanup = internalSetBreadcrumb(breadcrumb)
            const setters = createBreadcrumbSetters(depth + 1)
            return { unsubscribe: cleanup, ...setters }
        }

        return {
            useBreadcrumb,
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
    <nav className="d-flex p-2" aria-label="Breadcrumbs">
        {sortBy(breadcrumbs, 'depth')
            .map(({ breadcrumb }) => breadcrumb)
            .filter(isDefined)
            .map(breadcrumb => {
                const divider =
                    breadcrumb.divider === undefined ? <ChevronRightIcon className="icon-inline" /> : breadcrumb.divider
                return (
                    <span key={breadcrumb.key} className="text-muted d-flex align-items-center test-breadcrumb">
                        <span className="font-weight-semibold">{divider}</span>
                        {breadcrumb.element}
                    </span>
                )
            })}
    </nav>
)

/**
 * To be used in unit tests, it minimally fulfills the BreadcrumbSetters interface.
 */
export const NOOP_BREADCRUMB_SETTERS: BreadcrumbSetters = {
    setBreadcrumb: () => ({ ...NOOP_BREADCRUMB_SETTERS, unsubscribe: () => undefined }),
    useBreadcrumb: () => ({
        ...NOOP_BREADCRUMB_SETTERS,
    }),
}
