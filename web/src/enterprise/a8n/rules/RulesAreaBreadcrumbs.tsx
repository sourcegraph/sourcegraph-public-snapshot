import React from 'react'
import { Link } from 'react-router-dom'
import { RulesAreaContext } from './scope/ScopeRulesArea'

export interface BreadcrumbItem {
    /** The text of the item. */
    text: string

    /** The link URL, or undefined if the text is not a link. */
    to?: string
}

interface Props extends Pick<RulesAreaContext, 'rulesURL'> {
    scopeItem?: Required<BreadcrumbItem>
    activeItem?: BreadcrumbItem

    className?: string
}

/**
 * The breadcrumbs for the rules area.
 */
export const RulesAreaBreadcrumbs: React.FunctionComponent<Props> = ({
    scopeItem,
    activeItem,
    rulesURL,
    className = '',
}) => (
    <nav className={`d-flex align-items-center ${className}`} aria-label="breadcrumb">
        <ol className="breadcrumb">
            {scopeItem && (
                <li className="breadcrumb-item">
                    <Link to={scopeItem.to}>{scopeItem.text}</Link>
                </li>
            )}
            <li className="breadcrumb-item">
                <Link to={rulesURL}>Rules</Link>
            </li>
            {activeItem && (
                <li className="breadcrumb-item active font-weight-bold">
                    {activeItem.to ? <Link to={activeItem.to}>{activeItem.text}</Link> : activeItem.text}
                </li>
            )}
        </ol>
    </nav>
)
