import React, { type ButtonHTMLAttributes, forwardRef, useEffect, useState } from 'react'

import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import {
    Button,
    Combobox,
    ComboboxInput,
    ComboboxList,
    ComboboxOption,
    ComboboxOptionText,
    createRectangle,
    Flipping,
    Popover,
    PopoverContent,
    type PopoverOpenEvent,
    PopoverTrigger,
    Strategy,
    usePopoverContext,
} from '@sourcegraph/wildcard'

import type { UserSelectSearchFields } from '../../graphql-operations'

import { useUserSelectSearch } from './backend'

import styles from './UserSelect.module.scss'

const POPOVER_PADDING = createRectangle(0, 0, 5, 5)

export interface UserSelectProps {
    disabled?: boolean
    htmlID?: string
    initialUsername?: string
    onSelect: (user: UserSelectSearchFields | null) => void
}

export const UserSelect: React.FunctionComponent<UserSelectProps> = ({
    htmlID,
    onSelect,
    initialUsername,
    disabled = false,
}) => {
    const [isOpen, setOpen] = useState(false)

    const [selectedUser, setSelectedUser] = useState<UserSelectSearchFields>()

    const handleOpenChange = (event: PopoverOpenEvent): void => {
        setOpen(event.isOpen)
    }

    const handleSelect = (user: UserSelectSearchFields | undefined): void => {
        setSelectedUser(user)
        setOpen(false)
        onSelect(user || null)
    }

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger
                as={UserSelectButton}
                id={htmlID}
                title={selectedUser?.username ?? initialUsername}
                disabled={disabled}
            />

            <PopoverContent
                targetPadding={POPOVER_PADDING}
                flipping={Flipping.opposite}
                strategy={Strategy.Fixed}
                className="d-flex"
            >
                <UserSelectContent selectedUser={selectedUser} onSelect={handleSelect} />
            </PopoverContent>
        </Popover>
    )
}

export interface UserSelectContentProps {
    selectedUser: UserSelectSearchFields | undefined
    onSelect: (user: UserSelectSearchFields) => void
}

export const UserSelectContent: React.FunctionComponent<UserSelectContentProps> = ({ onSelect }) => {
    const [search, setSearch] = useState<string>('')

    const { data, loading, error } = useUserSelectSearch(search)

    const selectHandler = (username: string): void => {
        const user = data?.users.nodes.find(user => user.username === username)
        if (user) {
            onSelect(user)
        }
    }

    useEffect(() => {
        if (error) {
            // eslint-disable-next-line no-console
            console.error(error)
        }
    }, [error])

    const suggestions: UserSelectSearchFields[] = data?.users.nodes || []

    return (
        <Combobox openOnFocus={true} className={styles.combobox} onSelect={selectHandler}>
            <ComboboxInput
                value={search}
                autoFocus={true}
                spellCheck={false}
                placeholder="Search users"
                aria-label="Search users"
                inputClassName={styles.comboboxInput}
                className={styles.comboboxInputContainer}
                onChange={event => setSearch(event.target.value)}
                status={loading ? 'loading' : error ? 'error' : 'initial'}
            />

            <ComboboxList className={styles.comboboxList}>
                {suggestions.map((item, index) => (
                    <UserOption key={item.id} item={item} index={index} />
                ))}
            </ComboboxList>
        </Combobox>
    )
}

interface UserOptionProps {
    item: UserSelectSearchFields
    index: number
}

const UserOption: React.FunctionComponent<UserOptionProps> = ({ item, index }) => (
    <ComboboxOption value={item.username} index={index} className={styles.comboboxOption}>
        <UserAvatar user={item} inline={true} className="mr-1" />{' '}
        <span>
            <ComboboxOptionText />
        </span>
    </ComboboxOption>
)

interface UserSelectButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    title: string | undefined
}

const UserSelectButton = forwardRef<HTMLButtonElement, UserSelectButtonProps>(function UserSelectButton(props, ref) {
    const { title, className, ...attributes } = props
    const { isOpen } = usePopoverContext()

    const Icon = isOpen ? ChevronUpIcon : ChevronDownIcon

    return (
        <Button
            {...attributes}
            ref={ref}
            variant="secondary"
            outline={true}
            aria-label="Choose a user"
            className={classNames(className, styles.triggerButton)}
        >
            <span className={styles.triggerButtonText}>{title ?? 'No user'}</span>

            <Icon className={styles.triggerButtonIcon} />
        </Button>
    )
})
