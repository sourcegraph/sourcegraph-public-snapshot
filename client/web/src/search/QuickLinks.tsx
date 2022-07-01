import React from 'react'

import classNames from 'classnames'
import LinkIcon from 'mdi-react/LinkIcon'

import { QuickLink } from '@sourcegraph/shared/src/schema/settings.schema'
import { Link, Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './QuickLinks.module.scss'

interface Props {
    quickLinks: QuickLink[] | undefined

    className?: string
}

export const QuickLinks: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ quickLinks, className = '' }) =>
    quickLinks && quickLinks.length > 0 ? (
        <div className={className}>
            {quickLinks.map((quickLink, index) => (
                <small className={classNames('text-nowrap mr-2', styles.quicklink)} key={index}>
                    <Tooltip content={quickLink.description}>
                        <Link to={quickLink.url}>
                            <Icon aria-hidden={true} className="pr-1" as={LinkIcon} />
                            {quickLink.name}
                        </Link>
                    </Tooltip>
                </small>
            ))}
        </div>
    ) : null
