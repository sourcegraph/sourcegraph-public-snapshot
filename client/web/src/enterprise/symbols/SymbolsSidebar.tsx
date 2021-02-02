import React from 'react'
import { Link, NavLink } from 'react-router-dom'
import { gql } from '../../../../shared/src/graphql/graphql'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'
// import { SymbolsSidebarContainerSymbolFields } from '../../graphql-operations'

export interface SymbolsSidebarOptions {
    // containerSymbol: SymbolsSidebarContainerSymbolFields
    containerSymbol: any
}

interface Props extends SymbolsSidebarOptions {
    allSymbolsURL: string
    className?: string
}

export const SymbolsSidebar: React.FunctionComponent<Props> = ({ containerSymbol, allSymbolsURL, className = '' }) => (
    <div>I am the sidebar</div>
    // <nav className={className}>
    //     <header className="mb-2">
    //         <Link to={allSymbolsURL} className="d-block small p-2 pb-1 pl-3">
    //             &laquo; All symbols
    //         </Link>
    //         <h2 className="mb-0">
    //             <NavLink to={containerSymbol.url} className="d-flex align-items-center p-2" {...commonNavLinkProps}>
    //                 <SymbolIcon kind={containerSymbol.kind} className="mr-1 flex-shrink-0" />
    //                 <span className="text-truncate">{containerSymbol.text}</span>
    //             </NavLink>
    //         </h2>
    //     </header>

    //     {containerSymbol.children.nodes.length > 0 ? (
    //         <ItemList symbols={containerSymbol.children.nodes} itemClassName="pl-2 pr-3 py-1" level={0} />
    //     ) : (
    //         <p className="text-muted">No child symbols</p>
    //     )}
    // </nav>
)
