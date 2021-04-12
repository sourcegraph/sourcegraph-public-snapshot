import classNames from 'classnames'
import React from 'react'

import { LinkWithIcon } from '../components/LinkWithIcon'

import { BatchChangesIconNav } from './icons'

interface Props {
    isSourcegraphDotCom: boolean
    className?: string
}

/**
 * An item in {@link GlobalNavbar} that links to the batch changes area.
 */
export const BatchChangesNavItem: React.FunctionComponent<Props> = ({ isSourcegraphDotCom, className }) => (
    <LinkWithIcon
        to={isSourcegraphDotCom ? 'https://about.sourcegraph.com/batch-changes' : '/batch-changes'}
        text="Batch Changes"
        icon={BatchChangesIconNav}
        className={classNames('nav-link btn btn-link text-decoration-none', className)}
        activeClassName={isSourcegraphDotCom ? undefined : 'active'}
    />
)
