import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { CatalogEntityOwnerFields } from '../../../../graphql-operations'
import { PersonLink } from '../../../../person/PersonLink'
import { CatalogGroupIcon } from '../CatalogGroupIcon'

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
            <Link to={owner.url} className={classNames('d-inline-flex', 'align-items-center', className)}>
                <CatalogGroupIcon className="icon-inline text-muted mr-1" /> {owner.name}
            </Link>
        ) : (
            <span className={className}>Unknown</span>
        )
    ) : (
        <span className={className}>{blankIfNone ? '' : 'None'}</span>
    )
