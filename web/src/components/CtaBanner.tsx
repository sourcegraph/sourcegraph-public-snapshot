import * as React from 'react'
import classNames from 'classnames'

interface Props {
    className?: string
    icon: React.ReactNode
    title: string
    bodyText: string
    href: string
    linkText: string
}

export const CtaBanner = React.memo<Props>(({ icon, className, title, bodyText, href, linkText }) => (
    <div className={classNames('web-content cta-banner shadow d-flex flex-row card', className)}>
        <div className="cta-banner__icon-column d-flex flex-column align-items-center">{icon}</div>
        <div>
            <h3>{title}</h3>
            <p>{bodyText}</p>
            <a href={href} target="_blank" rel="noreferrer" className="btn btn-primary ga-cta-install-now">
                {linkText}
            </a>
        </div>
    </div>
))
