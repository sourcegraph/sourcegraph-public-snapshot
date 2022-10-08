import React from 'react'

import { UsageExample } from './UsageExample'

interface Props {
    examples: { repo: string; file: string }[]
}

export const UsageExampleList: React.FunctionComponent<Props> = ({ examples }) => (
    <ol>
        {examples.map((example, index) => (
            <li key={index}>
                <UsageExample repo={example.repo} file={example.file} />
            </li>
        ))}
    </ol>
)
