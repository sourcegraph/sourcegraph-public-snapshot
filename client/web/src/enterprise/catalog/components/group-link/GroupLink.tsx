import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { gql } from '@sourcegraph/http-client'

import { GroupLinkFields } from '../../../../graphql-operations'
import { CatalogGroupIcon } from '../CatalogGroupIcon'

export const GROUP_LINK_FRAGMENT = gql`
    fragment GroupLinkFields on Group {
        url
        name
    }
`

interface Props {
    group: GroupLinkFields
    className?: string
}

export const GroupLink: React.FunctionComponent<Props> = ({ group, className }) => (
    <Link to={group.url} className={classNames('align-items-center d-inline-flex', className)}>
        <CatalogGroupIcon className="icon-inline text-muted mr-1" /> {group.name}
    </Link>
)
