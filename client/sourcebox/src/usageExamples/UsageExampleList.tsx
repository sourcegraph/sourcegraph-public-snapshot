import React from 'react'

import { UsageExample, UsageExampleData } from './UsageExample'

interface Props {
    usageExamples: UsageExampleData[]
}

export const UsageExampleList: React.FunctionComponent<Props> = ({ usageExamples }) => (
    <ol>
        {usageExamples.map((usageExample, index) => (
            <li key={index}>
                <UsageExample usageExample={usageExample} />
            </li>
        ))}
    </ol>
)
