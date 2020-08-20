import React from 'react'

interface Props {
    title: React.ReactNode
    icon: React.ComponentType<{ className?: string }>
    actions?: React.ReactNode
}

export const PageHeader: React.FunctionComponent<Props> = ({ title, icon: Icon, actions }) => (
    <>
        <div className="d-flex flex-wrap align-items-center">
            <h1 className="flex-grow-1">
                <Icon className="icon-inline" /> {title}
            </h1>
            {actions}
        </div>
    </>
)
