import React, { useEffect, useState } from 'react'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import {
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxPopover,
    MultiComboboxList,
    MultiComboboxOption,
    ComboboxOptionText,
} from '@sourcegraph/wildcard'

import type { Scalars, TeamMemberUserSelectSearchFields } from '../../../graphql-operations'

import { useUserSelectSearch } from './backend'

import styles from './UserSelect.module.scss'

export interface UserSelectProps {
    id?: string
    className?: string
    disabled?: boolean
    setSelectedMembers: (val: Scalars['ID'][]) => void
}

export const UserSelect: React.FunctionComponent<UserSelectProps> = ({
    id,
    disabled,
    className,
    setSelectedMembers,
}) => {
    const [search, setSearch] = useState<string>('')
    const [selectedItems, setSelectedItems] = useState<TeamMemberUserSelectSearchFields[]>([])

    useEffect(() => {
        setSelectedMembers(selectedItems.map(item => item.id))
    }, [selectedItems, setSelectedMembers])

    const { data, loading, error } = useUserSelectSearch(search)

    useEffect(() => {
        if (error) {
            // eslint-disable-next-line no-console
            console.error(error)
        }
    }, [error])

    const suggestions: TeamMemberUserSelectSearchFields[] = data?.users.nodes || []

    const suggestionsWithExcludes = suggestions.filter(
        item => !selectedItems.find(selectedItem => selectedItem.id === item.id)
    )

    return (
        <MultiCombobox
            selectedItems={selectedItems}
            getItemKey={item => item.id}
            getItemName={item => item.username}
            onSelectedItemsChange={setSelectedItems}
            className={className}
        >
            <MultiComboboxInput
                value={search}
                placeholder="Search users"
                onChange={event => setSearch(event.target.value)}
                status={loading ? 'loading' : error ? 'error' : 'initial'}
                disabled={disabled}
                id={id}
            />

            <MultiComboboxPopover>
                <MultiComboboxList items={suggestionsWithExcludes}>
                    {items => items.map((item, index) => <UserOption key={item.id} item={item} index={index} />)}
                </MultiComboboxList>
            </MultiComboboxPopover>
        </MultiCombobox>
    )
}

interface UserOptionProps {
    item: TeamMemberUserSelectSearchFields
    index: number
}

const UserOption: React.FunctionComponent<UserOptionProps> = ({ item, index }) => (
    <MultiComboboxOption value={item.username} index={index} className={styles.userOption}>
        <UserAvatar inline={true} user={item} className="mr-2" /> <ComboboxOptionText />
    </MultiComboboxOption>
)
