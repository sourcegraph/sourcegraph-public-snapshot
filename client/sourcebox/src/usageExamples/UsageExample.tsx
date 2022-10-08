import React from 'react'

export interface UsageExampleData {
    repo: string
    file: string
    excerpts?: { content: string }[]
}

interface Props {
    usageExample: UsageExampleData
}

export const UsageExample: React.FunctionComponent<Props> = ({ usageExample: { repo, file, excerpts } }) => (
    <p>
        {repo}: {file} - {excerpts?.map(excerpt => excerpt.content).join(', ')}
    </p>
)
