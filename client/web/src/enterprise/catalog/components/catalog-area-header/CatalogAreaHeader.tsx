import classNames from 'classnames'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import styles from './CatalogAreaHeader.module.scss'

interface PathComponent {
    to?: string
    icon?: React.ComponentType<{ className?: string }>
    text?: React.ReactNode
}

interface Props {
    path: PathComponent[]
    nav?: React.ReactFragment
}

export const CatalogAreaHeader: React.FunctionComponent<Props> = ({ path, nav }) =>
    path.length > 0 ? (
        <header className={styles.container}>
            <nav className={styles.ancestors}>
                {path.slice(0, -1).map(({ to, icon, text }, index) => (
                    <React.Fragment key={index}>
                        <PathComponent to={to} icon={icon} text={text} />
                        <span className={styles.divider}>/</span>
                    </React.Fragment>
                ))}
            </nav>
            <h1 className={classNames('mr-2', styles.header)}>
                <PathComponent {...path[path.length - 1]} />
            </h1>
            {nav}
        </header>
    ) : null

const PathComponent: React.FunctionComponent<PathComponent> = ({ to, icon: Icon, text }) => (
    <LinkOrSpan to={to}>
        {Icon && <Icon className={classNames('icon-inline', styles.icon)} />}
        {text && <span className={styles.text}>{text}</span>}
    </LinkOrSpan>
)
