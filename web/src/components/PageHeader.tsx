import React from 'react'
import { Breadcrumbs } from './Breadcrumbs'

interface Props {
    title: string
    icon: React.ReactNode
    breadcrumbs: React.ReactNode[]
    actions?: React.ReactNode
    badge?: string // TODO: consider support for multiple badges
}

export const PageHeader: React.FunctionComponent<Props> = ({ title, icon, actions, badge, breadcrumbs }) => (
    <>
        <Breadcrumbs breadcrumbs={breadcrumbs} />
        <div className="d-flex align-items-center">
            <h1 className="flex-grow-1 text-nowrap">
                {icon} {title}
                {badge && (
                    <sup>
                        <span className="badge badge-info text-uppercase">{badge}</span>
                    </sup>
                )}
            </h1>
            {actions}
        </div>
    </>
)
