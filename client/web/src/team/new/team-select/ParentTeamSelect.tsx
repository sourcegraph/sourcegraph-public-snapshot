import React, { type ButtonHTMLAttributes, forwardRef, useEffect, useState } from 'react'

import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'

import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
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

import type { ParentTeamSelectSearchFields, Scalars } from '../../../graphql-operations'

import { useParentTeamSelectSearch } from './backend'

import styles from './ParentTeamSelect.module.scss'

const POPOVER_PADDING = createRectangle(0, 0, 5, 5)

export interface ParentTeamSelectProps {
    disabled: boolean
    id?: Scalars['ID']
    teamId?: string
    initial?: string
    onSelect: (id: string | null) => void
}

export const ParentTeamSelect: React.FunctionComponent<ParentTeamSelectProps> = ({
    id,
    teamId,
    onSelect,
    initial,
    disabled,
}) => {
    const [isOpen, setOpen] = useState(false)

    const [parentTeam, setParentTeam] = useState<ParentTeamSelectSearchFields>()

    const handleOpenChange = (event: PopoverOpenEvent): void => {
        setOpen(event.isOpen)
    }

    const handleSelect = (parentTeam: ParentTeamSelectSearchFields | undefined): void => {
        setParentTeam(parentTeam)
        setOpen(false)
        onSelect(parentTeam?.name || null)
    }

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger
                as={ParentTeamSelectButton}
                id={id}
                title={parentTeam?.name ?? initial}
                disabled={disabled}
            />

            <PopoverContent
                targetPadding={POPOVER_PADDING}
                flipping={Flipping.opposite}
                strategy={Strategy.Absolute}
                className="d-flex"
            >
                <ParentTeamSelectContent teamId={teamId} parentTeam={parentTeam} onSelect={handleSelect} />
            </PopoverContent>
        </Popover>
    )
}

export interface ParentTeamSelectContentProps {
    teamId?: Scalars['ID']
    parentTeam: ParentTeamSelectSearchFields | undefined
    onSelect: (parentTeam: ParentTeamSelectSearchFields) => void
}

export const ParentTeamSelectContent: React.FunctionComponent<ParentTeamSelectContentProps> = ({
    teamId,
    onSelect,
}) => {
    const [search, setSearch] = useState<string>('')

    const { data, loading, error } = useParentTeamSelectSearch(teamId || null, search)

    const selectHandler = (name: string): void => {
        const team = data?.teams.nodes.find(team => team.name === name)
        if (team) {
            onSelect(team)
        }
    }

    useEffect(() => {
        if (error) {
            // eslint-disable-next-line no-console
            console.error(error)
        }
    }, [error])

    const suggestions: ParentTeamSelectSearchFields[] = data?.teams.nodes || []

    return (
        <Combobox openOnFocus={true} className={styles.combobox} onSelect={selectHandler}>
            <ComboboxInput
                value={search}
                autoFocus={true}
                spellCheck={false}
                placeholder="Search teams"
                aria-label="Search teams"
                inputClassName={styles.comboboxInput}
                className={styles.comboboxInputContainer}
                onChange={event => setSearch(event.target.value)}
                status={loading ? 'loading' : error ? 'error' : 'initial'}
            />

            <ComboboxList className={styles.comboboxList}>
                {suggestions.map((item, index) => (
                    <TeamOption key={item.id} item={item} index={index} />
                ))}
            </ComboboxList>
        </Combobox>
    )
}

interface TeamOptionProps {
    item: ParentTeamSelectSearchFields
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

interface ParentTeamSelectButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    title: string | undefined
}

const ParentTeamSelectButton = forwardRef<HTMLButtonElement, ParentTeamSelectButtonProps>(
    function ParentTeamSelectButton(props, ref) {
        const { title, className, ...attributes } = props
        const { isOpen } = usePopoverContext()

        const Icon = isOpen ? ChevronUpIcon : ChevronDownIcon

        return (
            <Button
                {...attributes}
                ref={ref}
                variant="secondary"
                outline={true}
                aria-label="Choose a parent team"
                className={classNames(className, styles.triggerButton)}
            >
                <span className={styles.triggerButtonText}>{title ?? 'No parent team'}</span>

                <Icon className={styles.triggerButtonIcon} />
            </Button>
        )
    }
)
