import React from 'react'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'

/**
 * An extension card shown in {@link ExtensionsExploreSection}.
 */
export const ExtensionsExploreSectionExtensionCard: React.FunctionComponent<{
    extensionID: string
    description?: string
    url: string
    className?: string
}> = ({ extensionID: title, description = '', url, className = '' }) => (
    <LinkOrSpan to={url} className={className}>
        <h4 className="mb-0">{title}</h4>
        {description && <small>{description}</small>}
    </LinkOrSpan>
)
