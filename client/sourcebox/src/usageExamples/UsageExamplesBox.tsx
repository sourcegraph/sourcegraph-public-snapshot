import React, { useState } from 'react'

interface Props {
    collapsible?: boolean
}

export const UsageExamplesBox: React.FunctionComponent<Props> = ({ collapsible }) => {
    const [collapsed, setCollapsed] = useState(collapsible)

    return (
        <aside>
            <header>Usage examples</header>
            {!collapsed && (
                <ol>
                    <li>a</li>
                    <li>b</li>
                </ol>
            )}
        </aside>
    )
}
