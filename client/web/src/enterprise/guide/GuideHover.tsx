import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { GuideHoverFields } from '../../graphql-operations'

export const GuideHoverFieldsGQLFragment = gql`
    fragment GuideHoverFields on GuideInfo {
        hover {
            markdown {
                text
            }
        }
    }
`

interface Props {
    guideInfo: GuideHoverFields

    /** A fragment to display after the hover signature and before the documentation. */
    afterSignature?: React.ReactFragment

    className?: string
}

export const GuideHover: React.FunctionComponent<Props> = ({ guideInfo, afterSignature, className = '' }) => {
    const hoverParts = guideInfo.hover?.markdown.text.split('---', 2)
    const hoverSig = hoverParts?.[0]
    const hoverDocumentation = hoverParts?.[1]

    return (
        <>
            {hoverSig && (
                <Markdown
                    dangerousInnerHTML={renderMarkdown(hoverSig)}
                    className={`symbol-hover__signature ${className}`}
                />
            )}
            {afterSignature}
            {hoverDocumentation && (
                <Markdown dangerousInnerHTML={renderMarkdown(hoverDocumentation)} className={className} />
            )}
        </>
    )
}
