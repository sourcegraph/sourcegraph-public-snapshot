import * as React from 'react'
import classNames from 'classnames'

interface Props {
    className?: string
    icon: React.ReactNode
    title: string
    bodyText: string
    href: string
    linkText: string
    googleAnalytics?: boolean
}

export const CtaBanner = React.memo<Props>(function CtaBanner({
    icon,
    className,
    title,
    bodyText,
    href,
    linkText,
    googleAnalytics,
}) {
    return (
        <div className={classNames('web-content cta-banner shadow d-flex flex-row card', className)}>
            <div className="cta-banner__icon-column d-flex flex-column align-items-center">{icon}</div>
            <div>
                <h3>{title}</h3>
                <p>{bodyText}</p>
                <a
                    href={href}
                    target="_blank"
                    rel="noreferrer"
                    className={classNames('btn btn-primary', { 'ga-cta-install-now': googleAnalytics })}
                >
                    {linkText}
                </a>
            </div>
        </div>
    )
})
