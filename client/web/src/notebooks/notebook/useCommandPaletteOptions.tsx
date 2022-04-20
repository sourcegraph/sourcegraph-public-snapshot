import { ReactElement, useMemo } from 'react'

import CodeTagsIcon from 'mdi-react/CodeTagsIcon'
import FunctionIcon from 'mdi-react/FunctionIcon'
import LanguageMarkdownOutlineIcon from 'mdi-react/LanguageMarkdownOutlineIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'

import { SymbolKind } from '@sourcegraph/shared/src/schema'
import { Icon } from '@sourcegraph/wildcard'

import { BlockInput } from '..'
import { parseFileBlockInput } from '../serialize'

interface CommandPaletteOption {
    id: string
    label: string
    icon: ReactElement
    onSelect: () => void
}

export const EMPTY_FILE_BLOCK_INPUT = { repositoryName: '', revision: '', filePath: '', lineRange: null }
export const EMPTY_SYMBOL_BLOCK_INPUT = {
    repositoryName: '',
    revision: '',
    filePath: '',
    symbolName: '',
    symbolContainerName: '',
    symbolKind: SymbolKind.UNKNOWN,
    lineContext: 3,
}

interface UseCommandPaletteOptionsProps {
    input: string
    addBlock: (blockInput: BlockInput) => void
}

export const useCommandPaletteOptions = ({ input, addBlock }: UseCommandPaletteOptionsProps): CommandPaletteOption[] =>
    useMemo(() => {
        const trimmedInput = input.trimStart()
        if (trimmedInput.startsWith('/')) {
            const inputQuery = trimmedInput.slice(1)
            return [
                {
                    id: 'add-md-block',
                    label: 'Add Markdown text',
                    icon: <Icon as={LanguageMarkdownOutlineIcon} size="md" />,
                    onSelect: () => addBlock({ type: 'md', input: { text: '', initialFocusInput: true } }),
                },
                {
                    id: 'add-query-block',
                    label: 'Add a Sourcegraph query',
                    icon: <Icon as={MagnifyIcon} size="md" />,
                    onSelect: () => addBlock({ type: 'query', input: { query: '', initialFocusInput: true } }),
                },
                {
                    id: 'add-file-block',
                    label: 'Add code from a file',
                    icon: <Icon as={CodeTagsIcon} size="md" />,
                    onSelect: () => addBlock({ type: 'file', input: EMPTY_FILE_BLOCK_INPUT }),
                },
                {
                    id: 'add-symbol-block',
                    label: 'Add a symbol',
                    icon: <Icon as={FunctionIcon} size="md" />,
                    onSelect: () => addBlock({ type: 'symbol', input: EMPTY_SYMBOL_BLOCK_INPUT }),
                },
            ].filter(option => option.label.toLowerCase().includes(inputQuery))
        }

        const parsedFileBlockInput = parseFileBlockInput(input.trim())
        if (parsedFileBlockInput.repositoryName && parsedFileBlockInput.filePath) {
            return [
                {
                    id: 'add-file-from-url',
                    label: `Add code from ${parsedFileBlockInput.filePath}`,
                    icon: <Icon as={CodeTagsIcon} size="md" />,
                    onSelect: () => addBlock({ type: 'file', input: parsedFileBlockInput }),
                },
            ]
        }

        const inputSummary = trimmedInput.length < 64 ? trimmedInput : `${trimmedInput.slice(0, 64)}...`
        return [
            {
                id: 'add-md-block-with-input',
                label: `Add Markdown text "${inputSummary}"`,
                icon: <Icon as={LanguageMarkdownOutlineIcon} size="md" />,
                onSelect: () => addBlock({ type: 'md', input: { text: trimmedInput, initialFocusInput: true } }),
            },
            {
                id: 'add-query-block-with-input',
                label: `Add a Sourcegraph query "${inputSummary}"`,
                icon: <Icon as={MagnifyIcon} size="md" />,
                onSelect: () => addBlock({ type: 'query', input: { query: trimmedInput, initialFocusInput: true } }),
            },
            {
                id: 'find-files-with-input',
                label: `Find files matching "${inputSummary}"`,
                icon: <Icon as={CodeTagsIcon} size="md" />,
                onSelect: () =>
                    addBlock({ type: 'file', input: { ...EMPTY_FILE_BLOCK_INPUT, initialQueryInput: trimmedInput } }),
            },
            {
                id: 'find-symbols-with-input',
                label: `Find symbols matching "${inputSummary}"`,
                icon: <Icon as={FunctionIcon} size="md" />,
                onSelect: () =>
                    addBlock({
                        type: 'symbol',
                        input: { ...EMPTY_SYMBOL_BLOCK_INPUT, initialQueryInput: trimmedInput },
                    }),
            },
        ]
    }, [input, addBlock])
