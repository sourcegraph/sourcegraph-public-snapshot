import React from 'react'

import { Button, ButtonGroup } from '@sourcegraph/wildcard'

import { DiffMode } from './RepositoryCommitPage'

interface DiffModeSelectorProps {
    className?: string
    small?: boolean
    onHandleDiffMode: (mode: DiffMode) => void
    diffMode: DiffMode
}

export const DiffModeSelector: React.FunctionComponent<React.PropsWithChildren<DiffModeSelectorProps>> = ({
    className,
    diffMode,
    onHandleDiffMode,
    small,
}) => (
    <div className={className}>
        <ButtonGroup>
            <Button
                onClick={() => onHandleDiffMode('unified')}
                size={small ? 'sm' : undefined}
                variant="secondary"
                outline={diffMode !== 'unified'}
                className="test-diff-mode-selector-unified"
            >
                Unified
            </Button>
            <Button
                onClick={() => onHandleDiffMode('split')}
                size={small ? 'sm' : undefined}
                variant="secondary"
                outline={diffMode !== 'split'}
                className="test-diff-mode-selector-split"
            >
                Split
            </Button>
        </ButtonGroup>
    </div>
)
