import * as React from 'react'
import { SourcegraphIcon } from './Icons'

export interface Props {
    url?: string
    style?: React.CSSProperties
    iconStyle?: React.CSSProperties
    className?: string
    ariaLabel?: string
    onClick?: (e: React.MouseEvent<HTMLElement>) => void
    target?: string
    label: string
}

export const Button: React.FunctionComponent<Props> = (props: Props) => (
    <a
        href={props.url}
        aria-label={props.ariaLabel}
        className={props.className}
        style={props.style}
        onClick={props.onClick}
        target={props.target}
    >
        <SourcegraphIcon style={props.iconStyle || { marginTop: '-1px', paddingRight: '4px', fontSize: '18px', verticalAlign: 'text-top' }} />
        {props.label}
    </a>
)
