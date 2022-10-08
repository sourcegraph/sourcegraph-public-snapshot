import React from 'react'

interface Props {
    repo: string
    file: string
}

export const UsageExample: React.FunctionComponent<Props> = ({ repo, file }) => (
    <p>
        {repo}: {file}
    </p>
)
