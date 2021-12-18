import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { GroupDetailFields } from '../../../../graphql-operations'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'

interface Props {
    group: Pick<GroupDetailFields, 'url' | 'name'>
    className?: string
}

export const GroupLink: React.FunctionComponent<Props> = ({ group, className }) => (
    <Link to={group.url} className={classNames('align-items-center d-inline-flex', className)}>
        <CatalogGroupIcon className="icon-inline text-muted mr-1" /> {group.name}
    </Link>
)
