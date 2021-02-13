import React from 'react'
import { Symbol } from './SymbolPage'
import {
    DocSymbolFields,
    DocSymbolFieldsFragment,
    DocSymbolHierarchyFragment,
    RepositoryFields,
    SymbolPageSymbolFields,
} from '../../graphql-operations'

export interface SymbolsSidebarOptions {
    containerSymbol?: Symbol
}

interface Props {
    className?: string
    containerSymbol?: Symbol
    repo: RepositoryFields
}

export const SymbolsSidebar: React.FunctionComponent<Props> = ({ containerSymbol, repo, className = '' }) => (
    <div>
        {containerSymbol ? (
            <ul>
                <SymbolsHierarchy repo={repo} containerSymbol={containerSymbol} />
            </ul>
        ) : (
            <span>Loading</span>
        )}
    </div>

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

function urlForSymbol(symbol: Symbol, repo: RepositoryFields): string {
    // TODO(beyang): this is a hack
    return `/${repo.name}/-/docs/${symbol.id}`
}

interface SymbolsHierarchyProps {
    containerSymbol: Symbol
    repo: RepositoryFields
}

export const SymbolsHierarchy: React.FunctionComponent<SymbolsHierarchyProps> = ({ containerSymbol, repo }) => {
    return (
        <li>
            <a href={urlForSymbol(containerSymbol, repo)}>
                {containerSymbol?.kind} {containerSymbol?.text}
            </a>
            {containerSymbol?.children?.map(child => (
                <ul key={child.kind + ':' + child.id}>
                    <SymbolsHierarchy repo={repo} containerSymbol={child} />
                </ul>
            ))}
        </li>
    )
}
