import React from 'react'

import { WorkspaceRoot } from '@sourcegraph/extension-api-types'

interface FuzzyFinderResultProps {
    value: string
    onClick: () => void
    workspaceRoot: WorkspaceRoot | undefined
}

// Note: filenames don't have to be in monospace font.

export const FuzzyFinderResult: React.FC<FuzzyFinderResultProps> = ({ value, onClick, workspaceRoot }) => {
    console.log('TODO')

    if (!workspaceRoot) {
        return (
            <div>
                <h3>Nagivate to a repo to use fuzzy finder</h3>
            </div>
        )
    }

    // TODO: language icon by file extension

    return (
        <div>
            <h1>{value}</h1>
        </div>
    )
}
