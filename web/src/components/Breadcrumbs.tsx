import React, { useState, useCallback, useEffect } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Link } from '../../../shared/src/components/Link'

export interface Breadcrumb {
    key: string
    element: React.ReactNode | null
    divider?: React.ReactNode
}

export interface BreadcrumbsProps {
    breadcrumbs: Breadcrumb[]
}

export interface UpdateBreadcrumbsProps {
    setBreadcrumb: (options: Breadcrumb) => () => void
}

export const useBreadcrumbs = (): BreadcrumbsProps & UpdateBreadcrumbsProps => {
    const [breadcrumbs, setBreadcrumbs] = useState<Breadcrumb[]>([
        { key: 'Home', element: <Link to="/search">Home</Link>, divider: null },
    ])
    const setBreadcrumb = useCallback((breadcrumb: Breadcrumb) => {
        console.log('setBreadcrumb', breadcrumb)
        setBreadcrumbs(breadcrumbs => {
            const index = breadcrumbs.findIndex(({ key }) => breadcrumb.key === key)
            if (index === -1) {
                return [...breadcrumbs, breadcrumb]
            }
            return [...breadcrumbs.slice(0, index), breadcrumb, ...breadcrumbs.slice(index + 1)]
        })
        return () => {
            // Replace with null (but remember order in case the key gets set again)
            setBreadcrumb({ ...breadcrumb, element: null })
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
            .map(({ element, key, divider = <ChevronRightIcon className="icon-inline" /> }) => (
                <React.Fragment key={key}>
                    <span className="breadcrumbs__divider">{divider}</span>
                    {element}
                </React.Fragment>
            ))}
    </>
)
