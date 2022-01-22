import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { gql } from '@sourcegraph/http-client'

import { GroupLinkFields2 } from '../../../../graphql-operations'
import { CatalogGroupIcon } from '../CatalogGroupIcon'

// TODO(sqs): handle 2 fragments wanting to be called GroupLinkFields

export const GROUP_LINK_FRAGMENT = gql`
    fragment GroupLinkFields2 on Group {
        url
        name
    }
`

interface Props {
    group: GroupLinkFields2
    className?: string
}

export const GroupLink: React.FunctionComponent<Props> = ({ group, className }) => (
    <Link to={group.url} className={classNames('align-items-center d-inline-flex', className)}>
        <CatalogGroupIcon className="icon-inline text-muted mr-1" /> {group.name}
    </Link>
)
