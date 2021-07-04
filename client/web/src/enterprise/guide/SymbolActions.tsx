import React from 'react'
import { gql } from '../../../../shared/src/graphql/graphql'
import { SymbolActionsFields } from '../../graphql-operations'
import { Link } from 'react-router-dom'

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

export const SymbolActions: React.FunctionComponent<Props> = ({ symbol, className = '' }) =>
    symbol.definitions.nodes.length > 0 ? (
        <div className={`btn-group ${className}`}>
            <Link to={symbol.definitions.nodes[0].url} className="btn btn-secondary rounded-0">
                Go to definition
            </Link>
            <Link
                to={`${symbol.definitions.nodes[0].url}&tab=references` /* TODO(sqs): un-hardcode */}
                className="btn btn-secondary rounded-0"
            >
                Find references
            </Link>
        </div>
    ) : null
