import React from 'react'
import { Breadcrumbs, BreadcrumbsProps } from './Breadcrumbs'

interface Props extends BreadcrumbsProps {
    title: string
    icon: React.ReactNode
    label?: string
    actions?: React.ReactNode
    badge?: {
        label: string
        tooltip?: string
        type?: 'info' | 'danger'
    }
}

export const PageHeader: React.FunctionComponent<Props> = ({ title, icon, actions, badge, breadcrumbs, label }) => (
    <>
        <div className="ml-1">
            <Breadcrumbs breadcrumbs={breadcrumbs} />
        </div>
        <div className="d-flex container mt-4">
            <div className="h1">{icon}</div>
            <div className="flex-grow-1 ml-4">
                <h1 className="text-nowrap">
                    {title}{' '}
                    {badge && (
                        <sup>
                            <span
                                className={`badge badge-${badge.type ?? 'info'} text-uppercase`}
                                data-tooltip={badge?.tooltip}
                            >
                                {badge.label}
                            </span>
                        </sup>
                    )}
                </h1>
                {label}
            </div>
            <div>{actions}</div>
        </div>
    </>
)
