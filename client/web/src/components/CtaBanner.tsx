import * as React from 'react'

import classNames from 'classnames'

import { ButtonLink, Card, Text } from '@sourcegraph/wildcard'

import styles from './CtaBanner.module.scss'

interface Props {
    className?: string
    bodyTextClassName?: string
    icon: React.ReactNode
    title: React.ReactNode
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
            {title}
            <Text className={bodyTextClassName}>{bodyText}</Text>
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
