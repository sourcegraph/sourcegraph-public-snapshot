import React, { type ButtonHTMLAttributes, forwardRef, useEffect, useState } from 'react'

import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'

import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
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

import type { ParentTeamSelectSearchFields, UserSelectSearchFields } from '../../graphql-operations'
import { useUserSelectSearch } from '../../site-admin/user-select/backend'
import { useParentTeamSelectSearch } from '../../team/new/team-select/backend'

import styles from '../../site-admin/user-select/UserSelect.module.scss'

const POPOVER_PADDING = createRectangle(0, 0, 5, 5)

interface TeamSelectSearchFields extends ParentTeamSelectSearchFields {}

export interface UserTeamSelectProps {
    disabled?: boolean
    htmlID?: string
    initialUsername?: string
    onSelectUser: (user: UserSelectSearchFields | null) => void
    onSelectTeam: (user: TeamSelectSearchFields | null) => void
}

export const UserTeamSelect: React.FunctionComponent<UserTeamSelectProps> = ({
    htmlID,
    onSelectUser,
    onSelectTeam,
    initialUsername,
    disabled = false,
}) => {
    const [isOpen, setOpen] = useState(false)

    const [selectedUser, setSelectedUser] = useState<UserSelectSearchFields>()
    const [selectedTeam, setSelectedTeam] = useState<TeamSelectSearchFields>()

    const handleOpenChange = (event: PopoverOpenEvent): void => {
        setOpen(event.isOpen)
    }

    const handleUserSelect = (user: UserSelectSearchFields | undefined): void => {
        setSelectedUser(user)
        setOpen(false)
        onSelectUser(user || null)
    }

    const handleTeamSelect = (team: TeamSelectSearchFields | undefined): void => {
        setSelectedTeam(team)
        setOpen(false)
        onSelectTeam(team || null)
    }

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger
                as={UserSelectButton}
                id={htmlID}
                title={
                    selectedUser
                        ? selectedUser.username ?? initialUsername
                        : selectedTeam
                        ? selectedTeam.name
                        : undefined
                }
                disabled={disabled}
            />

            <PopoverContent
                targetPadding={POPOVER_PADDING}
                flipping={Flipping.opposite}
                strategy={Strategy.Fixed}
                className="d-flex"
            >
                <UserTeamSelectContent
                    selectedUser={selectedUser}
                    onSelectUser={handleUserSelect}
                    selectedTeam={selectedTeam}
                    onSelectTeam={handleTeamSelect}
                />
            </PopoverContent>
        </Popover>
    )
}

export interface UserTeamSelectContentProps {
    selectedUser: UserSelectSearchFields | undefined
    selectedTeam: TeamSelectSearchFields | undefined
    onSelectUser: (user: UserSelectSearchFields | undefined) => void
    onSelectTeam: (user: TeamSelectSearchFields | undefined) => void
}

export const UserTeamSelectContent: React.FunctionComponent<UserTeamSelectContentProps> = ({
    onSelectUser,
    onSelectTeam,
}) => {
    const [search, setSearch] = useState<string>('')

    const { data, loading, error } = useUserSelectSearch(search)

    const { data: teamData, loading: teamLoading, error: teamError } = useParentTeamSelectSearch(null, search)

    const selectHandler = (name: string): void => {
        const user = data?.users.nodes.find(user => user.username === name)
        if (user) {
            onSelectUser(user)
            onSelectTeam(undefined)
        }
        const team = teamData?.teams.nodes.find(team => team.name === name)
        if (team) {
            onSelectTeam(team)
            onSelectUser(undefined)
        }
    }

    useEffect(() => {
        if (error) {
            // eslint-disable-next-line no-console
            console.error(error)
        }
    }, [error])

    const userSuggestions: UserSelectSearchFields[] = data?.users.nodes || []
    const teamSuggestions: TeamSelectSearchFields[] = teamData?.teams.nodes || []

    return (
        <Combobox openOnFocus={true} className={styles.combobox} onSelect={selectHandler}>
            <ComboboxInput
                value={search}
                autoFocus={true}
                spellCheck={false}
                placeholder="Search users and teams"
                aria-label="Search users and teams"
                inputClassName={styles.comboboxInput}
                className={styles.comboboxInputContainer}
                onChange={event => setSearch(event.target.value)}
                status={loading || teamLoading ? 'loading' : error || teamError ? 'error' : 'initial'}
            />

            <ComboboxList className={styles.comboboxList}>
                {userSuggestions.map((item, index) => (
                    <UserOption key={item.id} item={item} index={index} />
                ))}
                {teamSuggestions.map((item, index) => (
                    <TeamOption key={item.id} item={item} index={index} />
                ))}
            </ComboboxList>
        </Combobox>
    )
}

interface UserTeamOptionProps {
    item: UserSelectSearchFields
    index: number
}

const UserOption: React.FunctionComponent<UserTeamOptionProps> = ({ item, index }) => (
    <ComboboxOption value={item.username} index={index} className={styles.comboboxOption}>
        <UserAvatar user={item} inline={true} className="mr-1" />{' '}
        <span>
            <ComboboxOptionText />
        </span>
    </ComboboxOption>
)

interface TeamOptionProps {
    item: TeamSelectSearchFields
    index: number
}

const TeamOption: React.FunctionComponent<TeamOptionProps> = ({ item, index }) => (
    <ComboboxOption value={item.name} index={index} className={styles.comboboxOption}>
        <TeamAvatar inline={true} team={item} className="mr-2" />{' '}
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
            aria-label="Choose a user or team"
            className={classNames(className, styles.triggerButton)}
        >
            <span className={styles.triggerButtonText}>{title ?? 'No user or team'}</span>

            <Icon className={styles.triggerButtonIcon} />
        </Button>
    )
})
