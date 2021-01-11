import React, { memo } from 'react'
import type { MdiReactIconProps } from 'mdi-react'

const CircleDashedIcon: React.FunctionComponent<MdiReactIconProps> = ({
    color = 'currentColor',
    size = 24,
    className = '',
    ...props
}) => (
    <svg {...props} className={className} width={size} height={size} fill={color} viewBox="0 0 24 24">
        <path d="M13 1.6c-.7-.07-1.4-.07-2.1 0l.1 1c.63-.06 1.3-.06 1.9 0l.1-1zm3.3.87l-.41.91c.58.26 1.1.58 1.6.95l.58-.81a8.69 8.69 0 00-1.8-1zm-8.6.01c-.64.3-1.2.64-1.8 1.1l.59.81c.51-.37 1.1-.69 1.6-.95l-.41-.9zm13 3.4l-.81.6c.37.5.69 1.1.95 1.6l.91-.42a8.97 8.97 0 00-1.1-1.8zm-17 .03c-.41.57-.76 1.2-1 1.8l.91.4c.26-.57.58-1.1.95-1.6l-.81-.58zm19 5.1l-.99.1a9.4 9.4 0 010 1.92l1 .1a10.34 10.34 0 00-.01-2.1zm-21 .03c-.07.7-.07 1.4 0 2.1l1-.1a9.65 9.65 0 010-1.9l-1-.1zm19 4.9c-.26.58-.58 1.1-.95 1.6l.81.59c.41-.57.76-1.2 1.1-1.8l-.91-.41zm-17 0l-.91.41c.29.64.64 1.2 1.1 1.8l.81-.59a12.9 12.9 0 01-.95-1.6zm14 3.8c-.51.37-1.1.7-1.6.95l.41.91c.64-.29 1.2-.64 1.8-1.1l-.59-.8zm-11 .01l-.59.81c.57.41 1.2.76 1.8 1l.41-.91a8.55 8.55 0 01-1.6-.95zm4.6 1.7l-.1 1c.7.07 1.4.07 2.1 0l-.1-1c-.63.07-1.3.07-1.9 0z" />
    </svg>
)

export default memo(CircleDashedIcon)
