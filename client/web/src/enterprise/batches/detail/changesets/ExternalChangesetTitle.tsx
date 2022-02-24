import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Icon } from '@sourcegraph/wildcard'

import { ExternalChangesetFields } from '../../../../graphql-operations'

interface Props extends Pick<ExternalChangesetFields, 'externalID' | 'externalURL' | 'title'> {
    /** Optionally, any class names to forward as a prop to the inner `LinkOrSpan` */
    className?: string
}

export const ExternalChangesetTitle: React.FunctionComponent<Props> = ({
    className,
    externalID,
    externalURL,
    title,
}) => (
    <LinkOrSpan to={externalURL?.url} target="_blank" rel="noopener noreferrer" className={className}>
        {title}
        {externalID && <> (#{externalID})</>}
        {externalURL?.url && (
            <>
                {' '}
                <Icon as={ExternalLinkIcon} inline={false} size="1rem" />
            </>
        )}
    </LinkOrSpan>
)
