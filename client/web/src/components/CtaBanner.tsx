import * as React from 'react'

import classNames from 'classnames'

import { ButtonLink, Card } from '@sourcegraph/wildcard'

import styles from './CtaBanner.module.scss'

interface Props {
    className?: string
    bodyTextClassName?: string
    icon: React.ReactNode
    headingElement?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6'
    title: string
    bodyText: string
    href: string
    linkText: string
    googleAnalytics?: boolean
    onClick?: () => void
}

export const CtaBanner: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    icon,
    className,
    bodyTextClassName,
    headingElement: HeadingX = 'h3',
    title,
    bodyText,
    href,
    linkText,
    googleAnalytics,
    onClick,
}) => (
    <Card className={classNames('shadow d-flex flex-row py-4 pr-4 pl-3', styles.ctaBanner, className)}>
        <div className="mr-4 d-flex flex-column align-items-center">{icon}</div>
        <div>
            <HeadingX>{title}</HeadingX>
            <p className={bodyTextClassName}>{bodyText}</p>
            <ButtonLink
                to={href}
                target="_blank"
                rel="noreferrer"
                onClick={onClick}
                className={classNames({ 'ga-cta-install-now': googleAnalytics })}
                variant="primary"
            >
                {linkText}
            </ButtonLink>
        </div>
    </Card>
)
