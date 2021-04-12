import React from 'react'

import { DiffMode } from './RepositoryCommitPage'

interface DiffModeSelectorProps {
    className?: string
    handleDiffMode: (mode: DiffMode) => void
    diffMode: DiffMode
}

export const DiffModeSelector: React.FunctionComponent<DiffModeSelectorProps> = ({
    className,
    diffMode,
    handleDiffMode,
}) => (
    <div className={className}>
        <div role="group" className="btn-group">
            <button
                onClick={() => handleDiffMode('unified')}
                type="button"
                className={`btn ${diffMode === 'unified' ? 'btn-secondary' : 'btn-outline-secondary'}`}
            >
                Unified
            </button>
            <button
                onClick={() => handleDiffMode('split')}
                type="button"
                className={`btn ${diffMode === 'split' ? 'btn-secondary' : 'btn-outline-secondary'}`}
            >
                Split
            </button>
        </div>
    </div>
)
