import React from 'react'
import { gql } from '../../../../shared/src/graphql/graphql'
import { SymbolActionsFields } from '../../graphql-operations'
import { Link } from 'react-router-dom'
import { toViewStateHashComponent } from '../../../../shared/src/util/url'

export const SymbolActionsGQLFragment = gql`
    fragment SymbolActionsFields on ExpSymbol {
        definitions {
            nodes {
                url
            }
        }
    }
`

interface Props {
    symbol: SymbolActionsFields

    className?: string
}

export const SymbolActions: React.FunctionComponent<Props> = ({ symbol, className = '' }) => {
    const rangeURL = symbol.definitions.nodes[0].url
    return symbol.definitions.nodes.length > 0 ? (
        <div className={`btn-group ${className}`}>
            <Link to={rangeURL} className="btn btn-secondary rounded-0">
                Go to definition
            </Link>
            <Link
                to={`${rangeURL + toViewStateHashComponent('references')}`}
                className="btn btn-secondary rounded-0"
            >
                Find references
            </Link>
        </div>
    ) : null
}
