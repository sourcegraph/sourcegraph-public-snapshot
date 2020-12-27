import React from 'react'
import { Link, NavLink } from 'react-router-dom'
import { gql } from '../../../../shared/src/graphql/graphql'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'
import { SymbolsSidebarContainerSymbolFields } from '../../graphql-operations'

export const SymbolsSidebarContainerSymbolGQLFragment = gql`
    fragment SymbolsSidebarContainerSymbolFields on ExpSymbol {
        ...SymbolsSidebarCommonSymbolFields
        children(filters: $filters) {
            nodes {
                ...SymbolsSidebarCommonSymbolFields
                children(filters: $filters) {
                    nodes {
                        ...SymbolsSidebarCommonSymbolFields
                    }
                }
            }
        }
    }

    fragment SymbolsSidebarCommonSymbolFields on ExpSymbol {
        text
        detail
        kind
        url
    }
`

type SymbolItem =
    | SymbolsSidebarContainerSymbolFields['children']['nodes'][0]
    | SymbolsSidebarContainerSymbolFields['children']['nodes'][0]['children']['nodes'][0]

export interface SymbolsSidebarOptions {
    containerSymbol: SymbolsSidebarContainerSymbolFields
}

const commonNavLinkProps: Pick<
    React.ComponentProps<NavLink>,
    'className' | 'style' | 'activeClassName' | 'activeStyle'
> = {
    style: { borderLeft: 'solid 3px transparent', borderLeftColor: 'transparent' },
    activeClassName: 'text-body',
    activeStyle: { borderLeftColor: 'var(--primary)' },
}

const Item: React.FunctionComponent<{
    symbol: SymbolItem
    level: number
    tag?: 'li'
    className?: string
}> = ({ symbol, level, tag: Tag = 'li', className = '' }) => (
    <Tag>
        <NavLink
            to={symbol.url}
            className={`d-flex align-items-center w-100 ${className}`}
            {...commonNavLinkProps}
            style={{
                ...commonNavLinkProps.style,
                fontSize: `${1 - 0.175 * level}rem`,
                borderLeftColor: level > 0 ? 'var(--secondary)' : 'transparent',
            }}
        >
            <SymbolIcon kind={symbol.kind} className="mr-1 flex-shrink-0 icon-inline" />
            <span className="text-truncate">{symbol.text}</span>
        </NavLink>
        {'children' in symbol && symbol.children.nodes.length > 0 && (
            <ItemList symbols={symbol.children.nodes} level={level + 1} itemClassName="pl-2 pr-3 py-1" />
        )}
    </Tag>
)

const ItemList: React.FunctionComponent<{
    symbols: SymbolItem[]
    level: number
    itemClassName?: string
}> = ({ symbols, level, itemClassName = '' }) => (
    <ul className="list-unstyled mb-2" style={{ marginLeft: `${level * 1.275}rem` }}>
        {symbols
            .sort((a, b) => (a.kind < b.kind ? -1 : 1))
            .map(symbol => (
                <Item key={symbol.url} symbol={symbol} className={itemClassName} level={level} />
            ))}
    </ul>
)

interface Props extends SymbolsSidebarOptions {
    allSymbolsURL: string
    className?: string
}

export const SymbolsSidebar: React.FunctionComponent<Props> = ({ containerSymbol, allSymbolsURL, className = '' }) => (
    <nav className={className}>
        <header className="mb-2">
            <Link to={allSymbolsURL} className="d-block small p-2 pb-1 pl-3">
                &laquo; All symbols
            </Link>
            <h2 className="mb-0">
                <NavLink to={containerSymbol.url} className="d-flex align-items-center p-2" {...commonNavLinkProps}>
                    <SymbolIcon kind={containerSymbol.kind} className="mr-1 flex-shrink-0" />
                    <span className="text-truncate">{containerSymbol.text}</span>
                </NavLink>
            </h2>
        </header>

        {containerSymbol.children.nodes.length > 0 ? (
            <ItemList symbols={containerSymbol.children.nodes} itemClassName="pl-2 pr-3 py-1" level={0} />
        ) : (
            <p className="text-muted">No child symbols</p>
        )}
    </nav>
)
