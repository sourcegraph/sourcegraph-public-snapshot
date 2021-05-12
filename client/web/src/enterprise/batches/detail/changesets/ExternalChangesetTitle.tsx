import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import { ExternalChangesetFields } from '../../../../graphql-operations'

export const ExternalChangesetTitle: React.FunctionComponent<
    Pick<ExternalChangesetFields, 'externalID' | 'externalURL' | 'title'> & {
        /** Optionally, any class names to forward as a prop to the inner `LinkOrSpan` */
        className?: string
    }
> = ({ className, externalID, externalURL, title }) => (
    <LinkOrSpan to={externalURL?.url} target="_blank" rel="noopener noreferrer" className={className}>
        {title}
        {externalID && <> (#{externalID}) </>}
        {externalURL?.url && (
            <>
                {' '}
                <ExternalLinkIcon size="1rem" />
            </>
        )}
    </LinkOrSpan>
)
