import React from 'react'

interface Props {
    /** The sidebar. */
    sidebar: React.ReactElement

    /** The position of the sidebar relative to the content. */
    sidebarPosition: 'left' | 'right'

    className?: string
}

/**
 * A container with a sidebar and content.
 */
export const WithSidebar: React.FunctionComponent<Props> = ({ sidebar, sidebarPosition, className = '', children }) => (
    <div className={`d-flex ${sidebarPosition === 'left' ? 'flex-row-reverse' : ''} overflow-hidden ${className}`}>
        <div className="flex-1 overflow-auto">{children}</div>
        {React.cloneElement(sidebar, { className: `${sidebar.props.className || ''} flex-0` })}
    </div>
)
