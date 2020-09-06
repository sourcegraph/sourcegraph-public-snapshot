import React from 'react'

interface Props {
    title: string
    value: string | null
    className?: string
}

export const Search2SidebarFacet: React.FunctionComponent<Props> = ({ title, value, className = '' }) => (
    <section className={`Search2SidebarFacet border-bottom ${className}`}>
        <header className="card-header border-0">{title}</header>
        <div className="card-body d-flex align-items-center justify-content-between">
            {value !== null ? <code>{value}</code> : <>&mdash;</>}
            <span className="btn btn-link btn-sm p-0">Edit</span>
        </div>
    </section>
)
