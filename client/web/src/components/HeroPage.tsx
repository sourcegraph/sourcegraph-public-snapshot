import * as React from 'react'

import classNames from 'classnames'

import { Typography, Link } from '@sourcegraph/wildcard'

import styles from './HeroPage.module.scss'

interface HeroPageProps {
    icon?: React.ComponentType<React.PropsWithChildren<any>>
    iconLinkTo?: string
    iconClassName?: string
    className?: string
    title?: string | JSX.Element
    subtitle?: string | JSX.Element
    detail?: React.ReactNode
    body?: React.ReactNode
    cta?: JSX.Element
    lessPadding?: boolean
}

export const HeroPage: React.FunctionComponent<React.PropsWithChildren<HeroPageProps>> = props => (
    <div
        className={classNames(
            styles.heroPage,
            props.lessPadding && styles.lessPadding,
            !props.lessPadding && styles.defaultPadding,
            props.className
        )}
    >
        {props.icon && (
            <div className={classNames(styles.icon, props.iconClassName)}>
                {props.iconLinkTo ? (
                    <Link to={props.iconLinkTo}>
                        <props.icon />
                    </Link>
                ) : (
                    <props.icon />
                )}
            </div>
        )}
        {props.title && <Typography.H1 className={styles.title}>{props.title}</Typography.H1>}
        {props.subtitle && (
            <div data-testid="hero-page-subtitle" className={styles.subtitle}>
                {props.subtitle}
            </div>
        )}
        {props.detail && <div>{props.detail}</div>}
        {props.body}
        {props.cta && <div className={styles.cta}>{props.cta}</div>}
    </div>
)
