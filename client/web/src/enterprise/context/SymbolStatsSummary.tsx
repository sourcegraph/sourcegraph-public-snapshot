import React from 'react'
import { gql } from '../../../../shared/src/graphql/graphql'
import { SymbolStatsSummaryFields } from '../../graphql-operations'
import { Link } from 'react-router-dom'

// TODO(sqs): this is dummy
export const SymbolStatsSummaryGQLFragment = gql`
    fragment SymbolStatsSummaryFields on ExpSymbol {
        definitions {
            nodes {
                url
            }
        }
    }
`

interface Props {
    symbol: SymbolStatsSummaryFields

    className?: string
}

export const SymbolStatsSummary: React.FunctionComponent<Props> = ({ symbol, className = '' }) =>
    symbol.definitions.nodes.length > 0 ? (
        <ul className={`list-inline mb-0 ${className}`}>
            <li className="list-inline-item mr-3">
                Edited 8 days ago by <strong>@sqs</strong>
            </li>
            <li className="list-inline-item">
                Code owner: <strong>@eric</strong>
            </li>
        </ul>
    ) : null
