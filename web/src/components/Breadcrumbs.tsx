import React, { useState, useCallback, useEffect } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

export interface Breadcrumb {
    key: string
    element: React.ReactNode | null
}

export interface BreadcrumbsProps {
    breadcrumbs: Breadcrumb[]
}

export interface UpdateBreadcrumbsProps {
    setBreadcrumb: (key: string, element: React.ReactNode) => () => void
}

export const useBreadcrumbs = (): BreadcrumbsProps & UpdateBreadcrumbsProps => {
    const [breadcrumbs, setBreadcrumbs] = useState<Breadcrumb[]>([])
    const setBreadcrumb = useCallback((key: string, element: React.ReactNode) => {
        console.log('setBreadcrumb', key, element)
        setBreadcrumbs(breadcrumbs => {
            const index = breadcrumbs.findIndex(breadcrumb => breadcrumb.key === key)
            if (index === -1) {
                return [...breadcrumbs, { key, element }]
            }
            return [...breadcrumbs.slice(0, index), { key, element }, ...breadcrumbs.slice(index + 1)]
        })
        return () => {
            // Replace with null (but remember order in case the key gets set again)
            setBreadcrumb(key, null)
        }
    }, [])
    useEffect(() => console.log(breadcrumbs), [breadcrumbs])
    return {
        breadcrumbs,
        setBreadcrumb,
    }
}

export const Breadcrumbs: React.FunctionComponent<BreadcrumbsProps> = ({ breadcrumbs }) => (
    <>
        {breadcrumbs
            .filter(({ element }) => element !== null)
            .map(({ element, key }, index) => (
                <React.Fragment key={key}>
                    {index !== 0 && <ChevronRightIcon />} {element}
                </React.Fragment>
            ))}
    </>
)
