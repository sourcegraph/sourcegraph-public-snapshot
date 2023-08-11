import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { Button, useOnClickOutside, Input } from '@sourcegraph/wildcard'

import type { BlockInput } from '..'

import { NotebookAddBlockButtons } from './NotebookAddBlockButtons'
import { useCommandPaletteOptions } from './useCommandPaletteOptions'

import styles from './NotebookCommandPaletteInput.module.scss'

interface NotebookCommandPaletteInputProps {
    index: number
    onAddBlock: (blockIndex: number, blockInput: BlockInput) => void
    onFocusPreviousBlock?: () => void
    onShouldDismiss?: () => void
    hasFocus?: boolean
}

export const NotebookCommandPaletteInput = React.forwardRef<HTMLInputElement, NotebookCommandPaletteInputProps>(
    ({ index, onAddBlock, onFocusPreviousBlock, onShouldDismiss: onDeselected, hasFocus }, reference) => {
        const [input, setInput] = useState('')
        const [selectedOptionId, setSelectedOptionId] = useState<string | null>(null)
        const [showCommandPalette, setShowCommandPalette] = useState(false)
        const rootReference = useRef<HTMLDivElement>(null)
        const inputReference = useRef<HTMLInputElement>(null)
        const mergedInputReference = useMergeRefs([inputReference, reference])

        const addBlock = useCallback(
            (blockInput: BlockInput) => {
                setInput('')
                inputReference.current?.blur()
                onAddBlock(index, blockInput)
            },
            [inputReference, index, onAddBlock]
        )

        const focusOption = useCallback(
            (id: string | null) => {
                setSelectedOptionId(id)
            },
            [setSelectedOptionId]
        )
        useEffect(() => {
            if (!selectedOptionId) {
                return
            }
            const optionButton = rootReference.current?.querySelector<HTMLButtonElement>(
                `[data-option-id="${selectedOptionId}"]`
            )
            optionButton?.focus()
        }, [selectedOptionId])

        const openCommandPalette = useCallback(() => {
            if (input.trim().length === 0) {
                return
            }
            setShowCommandPalette(true)
        }, [input, setShowCommandPalette])

        const closeCommandPalette = useCallback(() => {
            setShowCommandPalette(false)
            setSelectedOptionId(null)
        }, [setShowCommandPalette, setSelectedOptionId])

        const commandPaletteOptionsProps = useMemo(() => ({ input, addBlock }), [input, addBlock])
        const commandPaletteOptions = useCommandPaletteOptions(commandPaletteOptionsProps)

        const getNextOptionId = useCallback(() => {
            if (!selectedOptionId) {
                return commandPaletteOptions[0].id
            }
            const selectedOptionIndex = commandPaletteOptions.findIndex(option => option.id === selectedOptionId)
            return selectedOptionIndex === commandPaletteOptions.length - 1
                ? commandPaletteOptions[0].id
                : commandPaletteOptions[selectedOptionIndex + 1].id
        }, [commandPaletteOptions, selectedOptionId])

        const getPreviousOptionId = useCallback(() => {
            if (!selectedOptionId) {
                return null
            }
            const selectedOptionIndex = commandPaletteOptions.findIndex(option => option.id === selectedOptionId)
            return selectedOptionIndex === 0 ? null : commandPaletteOptions[selectedOptionIndex - 1].id
        }, [commandPaletteOptions, selectedOptionId])

        const onKeyDown = useCallback(
            (event: React.KeyboardEvent<HTMLInputElement>) => {
                if (event.key === 'ArrowDown') {
                    event.preventDefault() // Prevent page scroll
                    if (showCommandPalette) {
                        focusOption(getNextOptionId())
                    } else if (input.trim().length > 0) {
                        setShowCommandPalette(true)
                    }
                } else if (event.key === 'ArrowUp' && selectedOptionId === null) {
                    event.preventDefault() // Prevent page scroll
                    if (onFocusPreviousBlock) {
                        onFocusPreviousBlock()
                        closeCommandPalette()
                    }
                    if (input.trim().length === 0) {
                        onDeselected?.()
                    }
                } else if (event.key === 'ArrowUp') {
                    const previousOptionId = getPreviousOptionId()
                    focusOption(previousOptionId)
                    if (!previousOptionId) {
                        inputReference.current?.focus()
                    }
                    // Prevent page scroll
                    event.preventDefault()
                } else if (event.key === 'Escape') {
                    closeCommandPalette()
                    onDeselected?.()
                }
                // Stop other notebook event handlers reacting to the input.
                event.stopPropagation()
                event.nativeEvent.stopImmediatePropagation()
            },
            [
                selectedOptionId,
                showCommandPalette,
                focusOption,
                getNextOptionId,
                closeCommandPalette,
                input,
                onFocusPreviousBlock,
                getPreviousOptionId,
                onDeselected,
            ]
        )

        useEffect(() => {
            if (input.trim().length > 0) {
                openCommandPalette()
            } else {
                closeCommandPalette()
            }
        }, [input, openCommandPalette, closeCommandPalette])

        useOnClickOutside(rootReference, closeCommandPalette)
        useOnClickOutside(rootReference, () => {
            onDeselected?.()
        })

        useEffect(() => {
            if (hasFocus) {
                inputReference.current?.focus()
            }
        }, [hasFocus])

        return (
            <div className={styles.root} ref={rootReference} data-testid="notebook-command-palette">
                <div className={styles.inputRow}>
                    <Input
                        ref={mergedInputReference}
                        className="w-100"
                        inputClassName={styles.input}
                        value={input}
                        onKeyDown={onKeyDown}
                        onChange={event => setInput(event.target.value)}
                        placeholder="Type something to get started, paste a file URL, or use / to open the command palette"
                        aria-label="Type something to get started, paste a file URL, or use / to open the command palette"
                        onFocus={openCommandPalette}
                        data-testid="command-palette-input"
                    />
                    {!showCommandPalette && <NotebookAddBlockButtons index={index} onAddBlock={onAddBlock} />}
                </div>
                {showCommandPalette && commandPaletteOptions.length > 0 && (
                    <div className={styles.commandPalette} onKeyDown={onKeyDown} role="menu" tabIndex={-1}>
                        {commandPaletteOptions.map(option => (
                            <Button
                                key={option.id}
                                className={classNames(
                                    styles.commandPaletteButton,
                                    selectedOptionId === option.id && styles.commandPaletteButtonSelected
                                )}
                                onClick={option.onSelect}
                                onMouseOver={() => setSelectedOptionId(option.id)}
                                onFocus={() => focusOption(option.id)}
                                data-option-id={option.id}
                                role="menuitem"
                            >
                                <span className="mr-2">{option.icon}</span>
                                {option.label}
                            </Button>
                        ))}
                    </div>
                )}
            </div>
        )
    }
)

NotebookCommandPaletteInput.displayName = 'NotebookCommandPaletteInput'
