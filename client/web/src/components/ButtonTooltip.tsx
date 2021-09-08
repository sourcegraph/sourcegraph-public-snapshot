import React from 'react'

import styles from './InputTooltip.module.scss'

type ButtonProps = React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement>

/**
 * A wrapper around `button` that restores the hover tooltip capability even if the button is disabled.
 *
 * Disabled elements do not trigger mouse events on hover, and `Tooltip` relies on mouse events.
 *
 * All other props are passed to the `button` element.
 */
export const ButtonTooltip: React.FunctionComponent<
    ButtonProps & { tooltip?: string } & Required<Pick<ButtonProps, 'type'>>
> = ({ disabled, tooltip, type, ...props }) => (
    <div className={styles.container}>
        {disabled && tooltip ? <div className={styles.tooltip} data-tooltip={tooltip} /> : null}
        {/* This ESLint rule requires specifying the type with a static string or a trivial ternary expression only.
            The intent is to avoid undesirable page reloading behavior because the default value of `type` for HTML
            `button`s is "submit". However, because we're using TS and require `type` on this component's props,
            we can guarantee a valid type is provided and achieve the same result. Hence the disable comment. */}
        {/* eslint-disable-next-line react/button-has-type */}
        <button type={type} disabled={disabled} data-tooltip={disabled ? undefined : tooltip} {...props} />
    </div>
)
