import * as React from 'react'

interface Props {
    title: string
    ctaText: string
    ctaLink: string
    onClick?: () => void
}

export const Banner = (props: Props) => (
    <div className="banner">
        <div className="banner__contents">
            <div className="banner__contents-title">
                {props.title}{' '}
                <a href={props.ctaLink} className="banner__contents-cta" onClick={props.onClick}>
                    {props.ctaText}
                </a>.
            </div>
        </div>
    </div>
)
