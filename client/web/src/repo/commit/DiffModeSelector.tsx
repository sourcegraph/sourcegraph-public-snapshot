import React from 'react'

import { Button, ButtonGroup, Input } from '@sourcegraph/wildcard'

import { DiffMode } from './RepositoryCommitPage'

import styles from './DiffModeSelector.module.scss'

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
        <ButtonGroup>
            <Button
                size={small ? 'sm' : undefined}
                variant="secondary"
                outline={diffMode !== 'unified'}
                as="label"
                className={styles.button}
            >
                <Input
                    type="radio"
                    name="diff-mode"
                    value="unified"
                    checked={diffMode === 'unified'}
                    onChange={event => onHandleDiffMode(event.target.value as DiffMode)}
                    className="sr-only"
                />
                Unified
            </Button>
            <Button
                size={small ? 'sm' : undefined}
                variant="secondary"
                outline={diffMode !== 'split'}
                as="label"
                className={styles.button}
            >
                <Input
                    type="radio"
                    name="diff-mode"
                    value="split"
                    checked={diffMode === 'split'}
                    onChange={event => onHandleDiffMode(event.target.value as DiffMode)}
                    className="sr-only"
                />
                Split
            </Button>
        </ButtonGroup>
    </div>
)
