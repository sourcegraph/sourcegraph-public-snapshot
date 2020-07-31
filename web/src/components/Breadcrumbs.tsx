import React, { useState, useCallback } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

export interface Breadcrumb {
    key: string
    element: React.ReactNode
}

export interface BreadcrumbsProps {
    breadcrumbs: Breadcrumb[]
}

export interface UpdateBreadcrumbsProps {
    pushBreadcrumb: (key: string, element: React.ReactNode) => () => void
}

export const useBreadcrumbs = (): BreadcrumbsProps & UpdateBreadcrumbsProps => {
    const [breadcrumbs, setBreadcrumbs] = useState<Breadcrumb[]>([])
    const pushBreadcrumb = useCallback(
        (key: string, element: React.ReactNode) => {
            console.log('pushBreadcrumb', key)
            setBreadcrumbs(breadcrumbs => [...breadcrumbs, { key, element }])
            return () => {
                console.log('popBreadcrumb', key)
                setBreadcrumbs(breadcrumbs => breadcrumbs.filter(breadcrumb => breadcrumb.key !== key))
            }
        },
        [setBreadcrumbs]
    )
    return {
        breadcrumbs,
        pushBreadcrumb,
    }
}

export const Breadcrumbs: React.FunctionComponent<BreadcrumbsProps> = props => {
    const { breadcrumbs } = props
    return (
        <>
            {breadcrumbs.map(({ element, key }, index) => (
                <React.Fragment key={key}>
                    {index !== 0 && <ChevronRightIcon />} {element}
                </React.Fragment>
            ))}
        </>
    )
}
