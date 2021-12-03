import React from 'react'

import { CatalogEntityOwnerFields } from '../../../../graphql-operations'
import { PersonLink } from '../../../../person/PersonLink'

interface Props {
    owner: CatalogEntityOwnerFields['owner']
    className?: string
}

export const EntityOwner: React.FunctionComponent<Props> = ({ owner, className }) =>
    owner ? (
        owner.__typename === 'Person' ? (
            <PersonLink person={owner} className={className} />
        ) : owner.__typename === 'Group' ? (
            <span>{owner.name}</span>
        ) : (
            <span>Unknown</span>
        )
    ) : (
        <span>None</span>
    )
