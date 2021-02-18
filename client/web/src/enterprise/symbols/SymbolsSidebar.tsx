import React from 'react'
import { Symbol } from './SymbolPage'
import { RepositoryFields, SymbolKind, SymbolTag } from '../../graphql-operations'
import { NavLink } from 'react-router-dom'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

export interface SymbolsSidebarOptions {
    containerSymbol?: Symbol
}

interface Props {
    className?: string
    containerSymbol?: Symbol
    repo: RepositoryFields
}

export const SymbolsSidebar: React.FunctionComponent<Props> = ({ containerSymbol, repo, className = '' }) =>
    containerSymbol === null ? (
        <p className="p-3 text-muted h3">Not found</p>
    ) : containerSymbol === undefined ? (
        <LoadingSpinner className="m-3" />
    ) : (
        <ItemList level={0} symbols={[containerSymbol]} repo={repo} />
    )

export function urlForSymbol(symbol: SymbolItem, repo: RepositoryFields): string {
    // TODO(beyang): this is a hack
    return `/${repo.name}/-/docs/${symbol.id}`
}

interface ItemListProps {
    level: number
    symbols: SymbolItem[]
    repo: RepositoryFields
}

interface SymbolItem {
    id: string
    text: string
    detail: string
    kind: SymbolKind
    tags: Array<SymbolTag>
    children?: SymbolItem[]
}

export const ItemList: React.FunctionComponent<ItemListProps> = ({ level, symbols, repo }) => {
    return (
        <ul
            className="list-unstyled"
            style={{
                overflowY: 'auto',
                flex: '1 1',
                display: 'flex',
                flexDirection: 'column',
                minHeight: 0,
            }}
        >
            {symbols.map(symbol => (
                <React.Fragment key={symbol.kind + ':' + symbol.id}>
                    <li>
                        <NavLink
                            to={urlForSymbol(symbol, repo)}
                            className="d-flex align-items-center px-1"
                            style={{
                                ...(level === 0
                                    ? {
                                          backgroundColor: 'var(--secondary)',
                                          paddingTop: '0.25rem',
                                          paddingBottom: '0.25rem',
                                          fontWeight: 'bold',
                                          fontSize: '1.2rem',
                                      }
                                    : {
                                          fontSize: '0.825rem',
                                          marginLeft: level > 1 ? `${(level - 1) * 1.1}rem` : 0,
                                          borderLeftColor: level > 1 ? 'var(--secondary)' : 'transparent',
                                          borderLeftWidth: 3,
                                          borderLeftStyle: 'solid',
                                      }),
                            }}
                            activeStyle={{
                                borderLeftColor: 'var(--primary)',
                                borderLeftWidth: 3,
                                borderLeftStyle: 'solid',
                            }}
                            exact={true}
                        >
                            <span>
                                <SymbolIcon kind={symbol.kind} className={'icon-inline mr-1'} />{' '}
                            </span>
                            <span className="text-truncate" dir={level === 0 ? 'rtl' : 'ltr'}>
                                {symbol.text}
                            </span>
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
