import classNames from 'classnames'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

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
            <Button
                onClick={() => onHandleDiffMode('unified')}
                className={classNames(
                    diffMode === 'unified' ? 'btn-secondary' : 'btn-outline-secondary',
                    small && 'btn-sm'
                )}
            >
                Unified
            </Button>
            <Button
                onClick={() => onHandleDiffMode('split')}
                className={classNames(
                    diffMode === 'split' ? 'btn-secondary' : 'btn-outline-secondary',
                    small && 'btn-sm'
                )}
            >
                Split
            </Button>
        </div>
    </div>
)
