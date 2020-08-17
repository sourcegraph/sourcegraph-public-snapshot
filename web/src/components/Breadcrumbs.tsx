import React, { useState, useCallback, useEffect, useRef } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Link } from '../../../shared/src/components/Link'

export interface Breadcrumb {
    key: string
    element: React.ReactNode | null
    divider?: React.ReactNode
}

export interface BreadcrumbsProps {
    breadcrumbs: BreadcrumbTree
}

interface BreadcrumbSchema {
    [key: string]: BreadcrumbSchema | null | undefined
}

interface BreadcrumbTree {
    [key: string]: {
        breadcrumb: Breadcrumb
        children: BreadcrumbTree
    }
}

export interface UpdateBreadcrumbsProps {
    setBreadcrumb: (options: Breadcrumb) => () => void
}

export const useBreadcrumbs = (schema: BreadcrumbSchema): BreadcrumbsProps & UpdateBreadcrumbsProps => {
    // lazy initialize state to prevent building breadcrumbs every render. home is common to all pages
    const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbTree>(() => ({
        home: {
            breadcrumb: { key: 'home', element: <Link to="/search">Home</Link>, divider: null },
            children: schemaToBreadcrumbs(schema),
        },
    }))

    const setBreadcrumb = useCallback((breadcrumb: Breadcrumb) => {
        console.log('setBreadcrumb', breadcrumb)

        const { key } = breadcrumb

        setBreadcrumbs(breadcrumbs => {
            replaceBreadcrumb(key, breadcrumb, breadcrumbs)
            return { ...breadcrumbs }
        })

        return () => {
            // cleanup function. only need to re-create top-level object to render
            setBreadcrumbs(breadcrumbs => {
                replaceBreadcrumb(key, { key, element: null }, breadcrumbs)
                return { ...breadcrumbs }
            })
        }
    }, [])

    useEffect(() => console.log(breadcrumbs), [breadcrumbs])
    return {
        breadcrumbs,
        setBreadcrumb,
    }
}

/**
 * Traverses breadcrumb tree until key is found, then replaces breadcrumb
 */
function replaceBreadcrumb(key: string, breadcrumb: Breadcrumb, breadcrumbTree: BreadcrumbTree): void {
    // first, check if the key exists at this level
    const node = breadcrumbTree[key]

    if (node) {
        node.breadcrumb = breadcrumb
        return
    }

    // if not, check all of the children's subtrees
    for (const subtreeKey of Object.keys(breadcrumbTree)) {
        replaceBreadcrumb(key, breadcrumb, breadcrumbTree[subtreeKey].children)
    }
}

/**
 * Initializes the breadcrumb tree from breadcrumb schema
 */
function schemaToBreadcrumbs(schema: BreadcrumbSchema, breadcrumbs: BreadcrumbTree = {}): BreadcrumbTree {
    for (const key of Object.keys(schema)) {
        breadcrumbs[key] = { breadcrumb: { key, element: null }, children: {} }
        const childrenSchema = schema[key]
        if (childrenSchema) {
            schemaToBreadcrumbs(childrenSchema, breadcrumbs[key].children)
        }
    }

    return breadcrumbs
}

/**
 * Renders a list of breadcrumbs given a tree of breadcrumbs.
 * Assumes that each level of the breadcrumb tree has only one active node (element != null)
 */
function renderBreadcrumbs(breadcrumbs: BreadcrumbTree, reactNodes: React.ReactNode[]): void {
    for (const key of Object.keys(breadcrumbs)) {
        const { breadcrumb, children } = breadcrumbs[key]
        if (!breadcrumb.element) {
            continue
        }

        reactNodes.push(
            <span key={key} className="text-muted d-flex align-items-center">
                <span className="font-weight-semibold">
                    {breadcrumb.divider !== undefined ? (
                        breadcrumb.divider
                    ) : (
                        <ChevronRightIcon className="icon-inline" />
                    )}
                </span>
                {breadcrumb.element}
            </span>
        )
        renderBreadcrumbs(children, reactNodes)
        break
    }
}

export const Breadcrumbs: React.FunctionComponent<BreadcrumbsProps> = ({ breadcrumbs }) => {
    const renderedBreadcrumbs: React.ReactNode[] = []

    renderBreadcrumbs(breadcrumbs, renderedBreadcrumbs)

    return (
        <nav className="d-flex" aria-label="Breadcrumbs">
            {renderedBreadcrumbs}
        </nav>
    )
}
