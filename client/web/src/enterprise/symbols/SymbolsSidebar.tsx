import React from 'react'
import { Symbol } from './SymbolPage'
import {
    DocSymbolFields,
    DocSymbolFieldsFragment,
    DocSymbolHierarchyFragment,
    RepositoryFields,
    SymbolPageSymbolFields,
} from '../../graphql-operations'
import { NavLink } from 'react-router-dom'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'

export interface SymbolsSidebarOptions {
    containerSymbol?: Symbol
}

interface Props {
    className?: string
    containerSymbol?: Symbol
    repo: RepositoryFields
}

export const SymbolsSidebar: React.FunctionComponent<Props> = ({ containerSymbol, repo, className = '' }) => (
    <div>{containerSymbol ? <ItemList level={0} symbols={[containerSymbol]} repo={repo} /> : <span>Loading</span>}</div>
)

function urlForSymbol(symbol: Symbol, repo: RepositoryFields): string {
    // TODO(beyang): this is a hack
    return `/${repo.name}/-/docs/${symbol.id}`
}

interface ItemListProps {
    level: number
    symbols: Symbol[]
    repo: RepositoryFields
}

export const ItemList: React.FunctionComponent<ItemListProps> = ({ level, symbols, repo }) => {
    return (
        <ul className="list-unstyled">
            {symbols.map(symbol => (
                <React.Fragment key={symbol.kind + ':' + symbol.id}>
                    <li>
                        <NavLink
                            className="d-flex align-items-center w-100 pl-1"
                            to={urlForSymbol(symbol, repo)}
                            style={
                                level === 0
                                    ? {
                                          fontSize: '2rem',
                                      }
                                    : {
                                          fontSize: '0.825rem',
                                          marginLeft: level > 1 ? `${(level - 1) * 1.1}rem` : 0,
                                          borderLeftColor: level > 1 ? 'var(--secondary)' : 'transparent',
                                          borderLeftWidth: 3,
                                          borderLeftStyle: 'solid',
                                      }
                            }
                            activeStyle={{
                                borderLeftColor: 'var(--primary)',
                                borderLeftWidth: 3,
                                borderLeftStyle: 'solid',
                            }}
                            exact={true}
                        >
                            <SymbolIcon kind={symbol.kind} /> <span className="text-truncate">{symbol.text}</span>
                        </NavLink>
                    </li>
                    {symbol.children && symbol.children?.length > 0 && (
                        <li>
                            <ItemList level={level + 1} symbols={symbol.children} repo={repo} />
                        </li>
                    )}
                </React.Fragment>
            ))}
        </ul>
    )
}
