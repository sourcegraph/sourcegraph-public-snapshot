import React from 'react'

import { mdiOpenInNew } from '@mdi/js'

import { Icon, LinkOrSpan } from '@sourcegraph/wildcard'

import type { ExternalChangesetFields } from '../../../../graphql-operations'

interface Props extends Pick<ExternalChangesetFields, 'externalID' | 'externalURL' | 'title'> {
    /** Optionally, any class names to forward as a prop to the inner `LinkOrSpan` */
    className?: string
}

export const ExternalChangesetTitle: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
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
                <Icon svgPath={mdiOpenInNew} inline={false} aria-hidden={true} height="1rem" width="1rem" />
            </>
        )}
    </LinkOrSpan>
)
