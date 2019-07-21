import React from 'react'
import { Link } from 'react-router-dom'

export interface BreadcrumbItem {
    /** The text of the item. */
    text: string

    /** The link URL, or undefined if the text is not a link. */
    to?: string
}

interface Props {
    /** The breadcrumb items. */
    items: BreadcrumbItem[]

    className?: string
}

/**
 * Breadcrumbs that convey the current page's location in a hierarchy.
 */
export const Breadcrumbs: React.FunctionComponent<Props> = ({ items, className = '' }) => (
    <nav className={`d-flex align-items-center ${className}`} aria-label="breadcrumb">
        <ol className="breadcrumb">
            {items.map(({ text, to }, i) => (
                <li key={i} className={`breadcrumb-item ${i === items.length - 1 ? 'active font-weight-bold' : ''}`}>
                    {to ? <Link to={to}>{text}</Link> : text}
                </li>
            ))}
        </ol>
    </nav>
)
