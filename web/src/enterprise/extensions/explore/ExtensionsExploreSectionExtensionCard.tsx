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
    <LinkOrSpan to={url} className="extensions-explore-section__card__content">
        <div className="extensions-explore-section__card__content__body">
            <p className="extensions-explore-section__card__content__body__title">{title}</p>
            {description && <p className="extensions-explore-section__card__content__body__text">{description}</p>}
        </div>
    </LinkOrSpan>
)
