import * as React from 'react'

export interface HeroPageProps {
    icon?: React.ComponentType<any>
    className?: string
    title?: string | JSX.Element
    subtitle?: string | JSX.Element
    detail?: React.ReactNode
    body?: React.ReactNode
    cta?: JSX.Element
}

export const HeroPage: React.FunctionComponent<HeroPageProps> = props => (
    <div className={`hero-page ${props.className || ''}`}>
        {props.icon && (
            <div className="hero-page__icon">
                <props.icon />
            </div>
        )}
        {props.title && <div className="hero-page__title">{props.title}</div>}
        {props.subtitle && <div className="hero-page__subtitle">{props.subtitle}</div>}
        {props.detail && <div className="hero-page__detail">{props.detail}</div>}
        {props.body}
        {props.cta && <div className="hero-page__cta">{props.cta}</div>}
    </div>
)
