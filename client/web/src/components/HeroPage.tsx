import type { FC } from 'react'

import classNames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Link, H1, Icon } from '@sourcegraph/wildcard'

import styles from './HeroPage.module.scss'

interface HeroPageProps {
    icon?: React.ComponentType
    iconLinkTo?: string
    iconClassName?: string
    iconAriaLabel?: string
    className?: string
    title?: string | JSX.Element
    subtitle?: string | JSX.Element
    detail?: React.ReactNode
    body?: React.ReactNode
    cta?: JSX.Element
    lessPadding?: boolean
}

export const HeroPage: FC<HeroPageProps> = props => (
    <div
        className={classNames(
            styles.heroPage,
            props.lessPadding && styles.lessPadding,
            !props.lessPadding && styles.defaultPadding,
            props.className
        )}
    >
        {props.icon && (
            <div className={classNames(styles.iconWrapper, props.iconClassName)}>
                {props.iconLinkTo ? (
                    <Link to={props.iconLinkTo} aria-label={props.iconAriaLabel || props.iconLinkTo}>
                        <Icon className={styles.icon} as={props.icon} aria-hidden={true} />
                    </Link>
                ) : (
                    <Icon className={styles.icon} as={props.icon} aria-hidden={true} />
                )}
            </div>
        )}
        {props.title && <H1 className={styles.title}>{props.title}</H1>}
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

interface NotFoundPageProps {
    pageType: string
}

export const NotFoundPage: FC<NotFoundPageProps> = ({ pageType }) => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle={`Sorry, the requested ${pageType} page was not found.`}
    />
)
