import * as React from 'react'

import { SourcegraphIcon } from '@sourcegraph/wildcard'

export interface SourcegraphIconButtonProps
    extends Pick<
        JSX.IntrinsicElements['a'],
        'href' | 'title' | 'rel' | 'className' | 'onClick' | 'onFocus' | 'target'
    > {
    /** CSS class applied to the icon */
    iconClassName?: string
    /** Text label shown next to the button */
    label?: string
    /** aria-label attribute */
    ariaLabel?: string
    /** data-testid attribute */
    dataTestId?: string
}

export const SourcegraphIconButton: React.FunctionComponent<SourcegraphIconButtonProps> = ({
    iconClassName,
    label,
    ariaLabel,
    className,
    href,
    onClick,
    onFocus,
    rel,
    target,
    title,
    dataTestId,
}) => (
    <a
        href={href}
        className={className}
        target={target ?? '_blank'}
        rel={rel ?? 'noopener noreferrer'}
        title={title}
        aria-label={ariaLabel}
        onClick={onClick}
        onFocus={onFocus}
        data-testid={dataTestId}
    >
        <SourcegraphIcon className={iconClassName} /> {label}
    </a>
)
