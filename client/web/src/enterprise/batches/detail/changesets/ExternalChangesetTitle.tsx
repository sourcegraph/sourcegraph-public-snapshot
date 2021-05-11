import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import { ExternalChangesetFields } from '../../../../graphql-operations'

export const ExternalChangesetTitle: React.FunctionComponent<
    Pick<ExternalChangesetFields, 'externalURL'> & {
        /** Optionally, any class names to apply to the wrapping `h3` element */
        className?: string
        /** If provided, will render the original title (`children`) in ~strikethrough~
         * with `newTitle` to the side */
        newTitle?: string
    }
> = ({ children, className, externalURL, newTitle }) => {
    const linkOrSpan = (
        <LinkOrSpan
            to={externalURL?.url}
            target="_blank"
            rel="noopener noreferrer"
            className={`mr-2 ${newTitle ? 'text-muted' : ''}`}
        >
            {children}
            {externalURL?.url && (
                <>
                    {' '}
                    <ExternalLinkIcon size="1rem" />
                </>
            )}
        </LinkOrSpan>
    )

    if (newTitle) {
        return (
            <h3 className={className}>
                <del>{linkOrSpan}</del>
                {newTitle}
            </h3>
        )
    }
    return <h3 className={className}>{linkOrSpan}</h3>
}
