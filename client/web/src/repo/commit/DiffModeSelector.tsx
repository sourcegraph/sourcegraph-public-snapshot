import React from 'react'

import classNames from 'classnames'

import { Button, ButtonGroup, Input } from '@sourcegraph/wildcard'

import type { DiffMode } from './RepositoryCommitPage'

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
                className={classNames(styles.button, 'mb-0')}
                as="label"
                htmlFor="diff-mode-selector-unified"
            >
                <Input
                    type="radio"
                    name="diff-mode"
                    value="unified"
                    checked={diffMode === 'unified'}
                    onChange={event => onHandleDiffMode(event.target.value as DiffMode)}
                    className="sr-only"
                    id="diff-mode-selector-unified"
                />
                Unified
            </Button>
            <Button
                size={small ? 'sm' : undefined}
                variant="secondary"
                outline={diffMode !== 'split'}
                className={classNames(styles.button, 'mb-0')}
                as="label"
                htmlFor="diff-mode-selector-split"
            >
                <Input
                    type="radio"
                    name="diff-mode"
                    value="split"
                    checked={diffMode === 'split'}
                    onChange={event => onHandleDiffMode(event.target.value as DiffMode)}
                    className="sr-only"
                    id="diff-mode-selector-split"
                />
                Split
            </Button>
        </ButtonGroup>
    </div>
)
