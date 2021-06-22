import classNames from 'classnames'
import * as React from 'react'

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
}

export const CtaBanner: React.FunctionComponent<Props> = ({
    icon,
    className,
    bodyTextClassName,
    headingElement: HeadingX = 'h3',
    title,
    bodyText,
    href,
    linkText,
    googleAnalytics,
}) => (
    <div className={classNames('cta-banner shadow d-flex flex-row card py-4 pr-4 pl-3', className)}>
        <div className="mr-4 d-flex flex-column align-items-center">{icon}</div>
        <div>
            <HeadingX>{title}</HeadingX>
            <p className={bodyTextClassName}>{bodyText}</p>
            <a
                href={href}
                // eslint-disable-next-line react/jsx-no-target-blank
                target="_blank"
                rel="noreferrer"
                className={classNames('btn btn-primary', { 'ga-cta-install-now': googleAnalytics })}
            >
                {linkText}
            </a>
        </div>
    </div>
)
