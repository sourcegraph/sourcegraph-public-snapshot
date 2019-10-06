import React from 'react'

interface Props {
    tag?: 'div' | 'ul' | 'ol'
    className?: string
}

/**
 * A timeline component with a line running down the left-hand side that appears to connect the
 * children.
 */
export const Timeline: React.FunctionComponent<Props> = ({ tag: Tag = 'ul', className = '', children }) => (
    <Tag className={`timeline ${className}`}>
        <div className="timeline__line" />
        {children}
    </Tag>
)
