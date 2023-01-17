import { forwardRef, ReactElement, useCallback, useState } from 'react'

import { mdiSourceRepository } from '@mdi/js'
import { ErrorLike } from '@storybook/client-api'
import { identity } from 'lodash'
import { useMergeRefs } from 'use-callback-ref'

import {
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxPopover,
    MultiComboboxList,
    MultiComboboxOption,
    MultiComboboxOptionText,
    ForwardReferenceComponent,
    Icon,
    InputStatus,
    ErrorMessage,
    useDebounce,
} from '@sourcegraph/wildcard'

import { useRepoSuggestions } from './hooks/use-repo-suggestions'

import styles from './RepositoriesField.module.scss'

interface RepositoriesFieldProps {
    value: string[]
    description: string
    status?: InputStatus | `${InputStatus}`
    error?: ErrorLike | string
    onChange: (value: string[]) => void
}

/**
 * Renders MultiCombobox UI repositories input with async-resolved suggestions.
 */
export const RepositoriesField = forwardRef(function RepositoriesField(props, reference) {
    const { value, description, className, status, error, onChange, ...attributes } = props

    const inputRef = useMergeRefs([reference])
    const [search, setSearch] = useState('')
    const debouncedSearch = useDebounce(search, 500)

    const { suggestions, loading } = useRepoSuggestions({
        search: debouncedSearch,
        selectedRepositories: value,
    })

    const handleSelect = useCallback(
        (repositories: string[]) => {
            onChange(repositories)
        },
        [onChange]
    )

    return (
        <MultiCombobox
            selectedItems={value}
            getItemName={identity}
            getItemKey={identity}
            onSelectedItemsChange={handleSelect}
        >
            <MultiComboboxInput
                {...attributes}
                ref={inputRef}
                value={search}
                autoComplete="off"
                status={loading ? 'loading' : status}
                className={className}
                onChange={event => setSearch(event.target.value)}
            />

            {error && (
                <small role="alert" aria-live="polite" className={styles.errorMessage}>
                    <ErrorMessage error={error} />
                </small>
            )}

            {!error && <small className={styles.description}>{description}</small>}

            <MultiComboboxPopover>
                <MultiComboboxList items={suggestions}>
                    {items => items.map((item, index) => <RepositoryOption key={item} value={item} index={index} />)}
                </MultiComboboxList>
            </MultiComboboxPopover>
        </MultiCombobox>
    )
}) as ForwardReferenceComponent<'input', RepositoriesFieldProps>

interface RepositoryOptionProps {
    value: string
    index: number
}

function RepositoryOption(props: RepositoryOptionProps): ReactElement {
    const { value, index } = props

    return (
        <MultiComboboxOption value={value} index={index} className={styles.suggestionsListItem}>
            <Icon
                className="mr-1"
                svgPath={mdiSourceRepository}
                inline={false}
                aria-hidden={true}
                height="1rem"
                width="1rem"
            />
            <MultiComboboxOptionText />
        </MultiComboboxOption>
    )
}
