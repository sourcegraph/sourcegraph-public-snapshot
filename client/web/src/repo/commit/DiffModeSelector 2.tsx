import classNames from 'classnames'
import React from 'react'

import { DiffMode } from './RepositoryCommitPage'

interface DiffModeSelectorProps {
    className?: string
    small?: boolean
    onHandleDiffMode: (mode: DiffMode) => void
    diffMode: DiffMode
}

export const DiffModeSelector: React.FunctionComponent<DiffModeSelectorProps> = ({
    className,
    diffMode,
    onHandleDiffMode,
    small,
}) => (
    <div className={className}>
        <div role="group" className="btn-group">
            <button
                onClick={() => onHandleDiffMode('unified')}
                type="button"
                className={classNames(
                    'btn',
                    diffMode === 'unified' ? 'btn-secondary' : 'btn-outline-secondary',
                    small && 'btn-sm'
                )}
            >
                Unified
            </button>
            <button
                onClick={() => onHandleDiffMode('split')}
                type="button"
                className={classNames(
                    'btn',
                    diffMode === 'split' ? 'btn-secondary' : 'btn-outline-secondary',
                    small && 'btn-sm'
                )}
            >
                Split
            </button>
        </div>
    </div>
)
