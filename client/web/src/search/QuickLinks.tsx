import classNames from 'classnames'
import LinkIcon from 'mdi-react/LinkIcon'
import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import { QuickLink } from '../schema/settings.schema'

import styles from './QuickLinks.module.scss'

interface Props {
    quickLinks: QuickLink[] | undefined

    className?: string
}

export const QuickLinks: React.FunctionComponent<Props> = ({ quickLinks, className = '' }) =>
    quickLinks && quickLinks.length > 0 ? (
        <div className={className}>
            {quickLinks.map((quickLink, index) => (
                <small className={classNames('text-nowrap mr-2', styles.quicklink)} key={index}>
                    <Link to={quickLink.url} data-tooltip={quickLink.description}>
                        <LinkIcon className="icon-inline pr-1" />
                        {quickLink.name}
                    </Link>
                </small>
            ))}
        </div>
    ) : null
