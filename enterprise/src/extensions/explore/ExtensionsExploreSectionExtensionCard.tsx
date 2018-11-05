import React from 'react'
import { LinkOrSpan } from '../../../../packages/webapp/src/components/LinkOrSpan'

/**
 * An extension card shown in {@link ExtensionsExploreSection}.
 */
export const ExtensionsExploreSectionExtensionCard: React.SFC<{
    title: string
    description?: string | React.ReactFragment
    url?: string
}> = ({ title, description = '', url }) => (
    <LinkOrSpan
        to={url}
        className="card bg-secondary border-primary card-link text-white"
        // TODO(sqs,francis): Make these styles usable with class names.
        //
        // tslint:disable-next-line:jsx-ban-props
        style={{ backgroundImage: 'linear-gradient(116deg, #0c1e41, #171941)' }}
    >
        <div className="card-body">
            <h2 className="h6 font-weight-bold mb-0 text-truncate">{title}</h2>
            {description && <p className="card-text mt-1 small">{description}</p>}
        </div>
    </LinkOrSpan>
)
