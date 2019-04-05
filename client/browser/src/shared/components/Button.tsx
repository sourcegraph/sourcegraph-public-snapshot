import * as React from 'react'
import { SourcegraphIcon } from './Icons'

interface Props {
    url?: string
    className?: string
    iconClassName?: string
    ariaLabel?: string
    onClick?: (e: React.MouseEvent<HTMLElement>) => void
    target?: string
    label: string
}

export const SourcegraphIconButton: React.FunctionComponent<Props> = (props: Props) => (
    <a
        href={props.url}
        aria-label={props.ariaLabel}
        className={props.className}
        onClick={props.onClick}
        target={props.target}
    >
        <SourcegraphIcon className={props.iconClassName} /> {props.label}
    </a>
)
