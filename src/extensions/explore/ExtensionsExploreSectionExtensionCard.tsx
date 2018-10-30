import React from 'react'
import { LinkOrSpan } from '../../components/LinkOrSpan'

/**
 * An extension card shown in {@link ExtensionsExploreSection}.
 */
export const ExtensionsExploreSectionExtensionCard: React.SFC<{
    title: string
    description?: string
    url?: string
}> = ({ title, description = '', url }) => (
    <LinkOrSpan to={url} className="card bg-secondary border-primary card-link text-white">
        <div className="card-body">
            <h2 className="card-title h6 font-weight-bold">{title}</h2>
            {description && <p className="card-text mt-1">{description}</p>}
        </div>
    </LinkOrSpan>
)
