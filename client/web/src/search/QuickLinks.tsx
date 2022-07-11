import React from 'react'

import { mdiLink } from '@mdi/js'
import classNames from 'classnames'

import { QuickLink } from '@sourcegraph/shared/src/schema/settings.schema'
import { Link, Icon } from '@sourcegraph/wildcard'

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
                    <Link to={quickLink.url} data-tooltip={quickLink.description}>
                        <Icon aria-hidden={true} className="pr-1" svgPath={mdiLink} />
                        {quickLink.name}
                    </Link>
                </small>
            ))}
        </div>
    ) : null
