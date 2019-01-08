import React from 'react'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'

/**
 * An extension card shown in {@link ExtensionsExploreSection}.
 */
export const ExtensionsExploreSectionExtensionCard: React.FunctionComponent<{
    extensionID: string
    description?: string
    url?: string
}> = ({ extensionID: title, description = '', url }) => (
    <LinkOrSpan
        to={url}
        className="extensions-explore-section--card--content"
    >
        <div className="extensions-explore-section--card--content--body">
            <p className="extensions-explore-section--card--content--body--title">{title}</p>
            {description && <p className="extensions-explore-section--card--content--body--text">{description}</p>}
        </div>
    </LinkOrSpan>
)
