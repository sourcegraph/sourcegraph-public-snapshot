import React from 'react'
import { Link } from 'react-router-dom'

import { CatalogEntityOwnerFields } from '../../../../graphql-operations'
import { PersonLink } from '../../../../person/PersonLink'

interface Props {
    owner: CatalogEntityOwnerFields['owner']
    blankIfNone?: string
    className?: string
}

export const EntityOwner: React.FunctionComponent<Props> = ({ owner, blankIfNone, className }) =>
    owner ? (
        owner.__typename === 'Person' ? (
            <PersonLink person={owner} className={className} />
        ) : owner.__typename === 'Group' ? (
            <Link to={owner.url} className={className}>
                {owner.name}
            </Link>
        ) : (
            <span className={className}>Unknown</span>
        )
    ) : (
        <span className={className}>{blankIfNone ? '' : 'None'}</span>
    )
