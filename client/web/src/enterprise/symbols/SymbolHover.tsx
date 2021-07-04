import React from 'react'
import H from 'history'
import { gql } from '../../../../shared/src/graphql/graphql'
import { SymbolHoverFields } from '../../graphql-operations'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../shared/src/util/markdown'

export const SymbolHoverGQLFragment = gql`
    fragment SymbolHoverFields on ExpSymbol {
        hover {
            markdown {
                text
            }
        }
    }
`

interface Props {
    symbol: SymbolHoverFields

    /** A fragment to display after the hover signature and before the documentation. */
    afterSignature?: React.ReactFragment

    className?: string

    history: H.History
    location: H.Location
}

export const SymbolHover: React.FunctionComponent<Props> = ({ symbol, afterSignature, className = '', history }) => {
    const hoverParts = symbol.hover?.markdown.text.split('---', 2)
    const hoverSig = hoverParts?.[0]
    const hoverDocumentation = hoverParts?.[1]

    return (
        <>
            {hoverSig && (
                <Markdown
                    dangerousInnerHTML={renderMarkdown(hoverSig)}
                    history={history}
                    className={`symbol-hover__signature ${className}`}
                />
            )}
            {afterSignature}
            {hoverDocumentation && (
                <Markdown
                    dangerousInnerHTML={renderMarkdown(hoverDocumentation)}
                    history={history}
                    className={className}
                />
            )}
        </>
    )
}
