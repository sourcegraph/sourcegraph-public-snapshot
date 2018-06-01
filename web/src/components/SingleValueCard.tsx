import * as React from 'react'
import { Link } from 'react-router-dom'

/**
 * A card that displays a large single value.
 */
export const SingleValueCard: React.SFC<{
    title: string
    subTitle?: string
    value: string | number
    link?: string
    className?: string
    valueClassName?: string
    valueTooltip?: string
}> = ({ title, value, subTitle, link, className, valueClassName, valueTooltip }) => {
    const v = link !== undefined ? <Link to={link}>{value}</Link> : <>{value}</>

    return (
        <div className={`card m-2 single-value-card ${className || ''}`}>
            <div className="card-body text-center">
                <h4 className="card-title mb-0">{title}</h4>
                <small className="card-text">{subTitle || ''}</small>
                <p
                    data-tooltip={valueTooltip}
                    className={`card-text font-weight-bold text-nowrap single-value-card__value ${valueClassName ||
                        ''}`}
                >
                    {v}
                </p>
            </div>
        </div>
    )
}
