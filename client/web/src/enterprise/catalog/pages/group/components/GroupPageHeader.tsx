import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import styles from './GroupPageHeader.module.scss'

interface PathComponent {
    to?: string
    icon?: React.ComponentType<{ className?: string }>
    text?: React.ReactNode
}

interface Props {
    path: PathComponent[]
}

export const GroupPageHeader: React.FunctionComponent<Props> = ({ path }) =>
    path.length > 0 ? (
        <header className={styles.container}>
            <div className="d-flex flex-wrap w-100">
                <ComponentAncestorsPath path={path.slice(0, -1)} className={styles.ancestors} />
                <h1 className={classNames('mr-2', styles.header)}>
                    <PathComponent {...path[path.length - 1]} />
                </h1>
            </div>
        </header>
    ) : null

const PathComponent: React.FunctionComponent<PathComponent> = ({ to, icon: Icon, text }) => (
    <LinkOrSpan to={to}>
        {Icon && <Icon className={classNames('icon-inline', styles.icon)} />}
        {text && <span className={styles.text}>{text}</span>}
    </LinkOrSpan>
)

const ComponentAncestorsPath: React.FunctionComponent<
    Pick<Props, 'path'> & {
        divider?: '/' | '>'
        className?: string
        componentClassName?: string
        lastComponentClassName?: string
    }
> = ({ path, divider = '/', className, componentClassName, lastComponentClassName }) => (
    <nav className={className}>
        {path.map(({ to, icon, text }, index) => (
            <span
                key={index}
                className={classNames(
                    'text-nowrap',
                    componentClassName,
                    index === path.length - 1 ? lastComponentClassName : undefined
                )}
            >
                <PathComponent to={to} icon={icon} text={text} />
                {divider === '>' ? (
                    <ChevronRightIcon className="icon-inline text-muted" />
                ) : (
                    <span className={styles.divider}>{divider}</span>
                )}
            </span>
        ))}
    </nav>
)
