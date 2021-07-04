import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'
import { renderMarkdown } from '../../../../shared/src/util/markdown'
import { ContainerSymbolsListSymbolFields } from '../../graphql-operations'
import { symbolHoverSynopsisMarkdown } from './symbolInfo'
import { gql } from '../../../../shared/src/graphql/graphql'

export const ContainerSymbolsListSymbolGQLFragment = gql`
    fragment ContainerSymbolsListSymbolFields on ExpSymbol {
        text
        detail
        kind
        url
        hover {
            markdown {
                text
            }
        }
    }
`

const Item: React.FunctionComponent<{
    symbol: ContainerSymbolsListSymbolFields
    tag?: 'li'
    className?: string
    history: H.History
}> = ({ symbol, tag: Tag = 'li', className = '', history }) => {
    const detailIncludesText = Boolean(symbol.detail?.includes(symbol.text))
    const title = detailIncludesText ? symbol.detail : symbol.text
    const subtitle = detailIncludesText ? null : symbol.detail

    const synopsisMarkdown = symbol.hover && symbolHoverSynopsisMarkdown(symbol.hover.markdown.text)
    return (
        <Tag className={`${className} d-flex`}>
            <Link to={symbol.url}>
                <SymbolIcon kind={symbol.kind} className="mr-2 h2 mb-0" />
            </Link>
            <div>
                <h3 className="mb-0">
                    <Link to={symbol.url}>{title}</Link>
                </h3>
                {subtitle && <p className="text-muted mb-0">{subtitle}</p>}
                {synopsisMarkdown && (
                    <Markdown dangerousInnerHTML={renderMarkdown(synopsisMarkdown)} history={history} />
                )}
            </div>
        </Tag>
    )
}

interface Props {
    symbols: ContainerSymbolsListSymbolFields[]
    history: H.History
}

export const ContainerSymbolsList: React.FunctionComponent<Props> = ({ symbols, history }) => (
    <div>
        <ul className="list-group list-group-flush border-bottom">
            {symbols
                .sort((a, b) => (a.detail || '').localeCompare(b.detail || ''))
                .map(symbol => (
                    <Item key={symbol.url} symbol={symbol} className="list-group-item p-2" history={history} />
                ))}
        </ul>
    </div>
)
