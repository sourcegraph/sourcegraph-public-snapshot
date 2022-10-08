import React, { useState } from 'react'

interface Props {
    query: string
    collapsible?: boolean
    theme?: 'dark' | 'light'
}

export const UsageExamplesBox: React.FunctionComponent<Props> = ({ query, collapsible, theme }) => {
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
