import classNames from 'classnames'
import React from 'react'

import { DiffMode } from './RepositoryCommitPage'

interface DiffModeSelectorProps {
    className?: string
    size?: 'small'
    handleDiffMode: (mode: DiffMode) => void
    diffMode: DiffMode
}

export const DiffModeSelector: React.FunctionComponent<DiffModeSelectorProps> = ({
    className,
    diffMode,
    handleDiffMode,
    size,
}) => (
    <div className={className}>
        <div role="group" className="btn-group">
            <button
                onClick={() => handleDiffMode('unified')}
                type="button"
                className={classNames(
                    'btn',
                    diffMode === 'unified' ? 'btn-secondary' : 'btn-outline-secondary',
                    size === 'small' && 'btn-sm'
                )}
            >
                Unified
            </button>
            <button
                onClick={() => handleDiffMode('split')}
                type="button"
                className={classNames(
                    'btn',
                    diffMode === 'split' ? 'btn-secondary' : 'btn-outline-secondary',
                    size === 'small' && 'btn-sm'
                )}
            >
                Split
            </button>
        </div>
    </div>
)
