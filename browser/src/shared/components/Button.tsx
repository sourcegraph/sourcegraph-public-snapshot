import * as React from 'react'
import { SourcegraphIcon } from './Icons'

export interface SourcegraphIconButtonProps
    extends Pick<
        JSX.IntrinsicElements['a'],
        'href' | 'title' | 'rel' | 'className' | 'onClick' | 'target' | 'aria-label'
    > {
    iconClassName?: string
    label?: string
}

export const SourcegraphIconButton: React.FunctionComponent<SourcegraphIconButtonProps> = ({
    iconClassName,
    label,
    ...anchorProps
}) => (
    <a {...anchorProps}>
        <SourcegraphIcon className={iconClassName} /> {label}
    </a>
)
