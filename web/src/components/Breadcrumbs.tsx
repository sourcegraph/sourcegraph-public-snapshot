import React, { useState } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

interface BreadcrumbsProps {
    breadcrumbs: React.ReactNode[]
}

export interface UpdateBreadcrumbsProps {
    pushBreadcrumb: (element: React.ReactNode) => () => void
}

export const useBreadcrumbs = (): BreadcrumbsProps & UpdateBreadcrumbsProps => {
    const [breadcrumbs, setBreadcrumbs] = useState<React.ReactNode[]>([])
    return {
        breadcrumbs,
        pushBreadcrumb: element => {
            setBreadcrumbs([...breadcrumbs, element])
            return () => setBreadcrumbs(breadcrumbs.filter(breadcrumb => breadcrumb !== element))
        },
    }
}

export const Breadcrumbs: React.FunctionComponent<BreadcrumbsProps> = props => {
    const { breadcrumbs } = props
    return (
        <>
            {breadcrumbs.map((breadcrumb, index) => (
                <>
                    {index !== 0 && <ChevronRightIcon />} {breadcrumb}
                </>
            ))}
        </>
    )
}
