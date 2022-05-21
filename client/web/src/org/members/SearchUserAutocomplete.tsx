import React, { useCallback, useEffect, useRef, useState } from 'react'

import { useLazyQuery } from '@apollo/client'
import classNames from 'classnames'
import { debounce } from 'lodash'
// eslint-disable-next-line no-restricted-imports
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { FormInput } from '@sourcegraph/wildcard'

import { AutocompleteMembersSearchResult, AutocompleteMembersSearchVariables, Maybe } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'

import { SEARCH_USERS_AUTOCOMPLETE_QUERY } from './gqlQueries'

import styles from './SearchUserAutocomplete.module.scss'

interface IUserItem {
    id: string
    username: string
    inOrg: boolean
    displayName: Maybe<string>
    avatarURL: Maybe<string>
}

interface AutocompleteSearchUsersProps {
    disabled?: boolean
    onValueChanged: (value: string, isEmail: boolean) => void
    orgId: string
}

const UserResultItem: React.FunctionComponent<
    React.PropsWithChildren<{
        onSelectUser: (user: IUserItem) => void
        onKeyDown: (key: string, index: number) => void
        index: number
        user: IUserItem
    }>
> = ({ user, onSelectUser, onKeyDown, index }) => {
    const selectUser = useCallback(() => {
        onSelectUser(user)
    }, [onSelectUser, user])

    const keyDown = useCallback(
        (event: React.KeyboardEvent) => {
            onKeyDown(event.key, index)
        },
        [onKeyDown, index]
    )

    return (
        <DropdownItem
            data-testid="search-context-menu-item"
            data-res-user-id={user.id}
            className={styles.item}
            onClick={selectUser}
            role="menuitem"
            disabled={user.inOrg}
            onKeyDown={keyDown}
        >
            <div className={classNames('d-flex align-items-center justify-content-between', styles.userContainer)}>
                <div className={styles.avatarContainer}>
                    <UserAvatar
                        size={24}
                        className={classNames(styles.avatar, user.inOrg ? styles.avatarDisabled : undefined)}
                        user={user}
                        data-tooltip={user.displayName || user.username}
                    />
                </div>
                <div className="d-flex flex-column">
                    <div>
                        <strong>{user.displayName || user.username}</strong>{' '}
                        {user.displayName && <span className={styles.userName}>{user.username}</span>}
                    </div>
                    {user.inOrg && <small className="text-muted">Already in this organization</small>}
                </div>
            </div>
        </DropdownItem>
    )
}

const EmptyResultsItem: React.FunctionComponent<
    React.PropsWithChildren<{
        userNameOrEmail: string
    }>
> = ({ userNameOrEmail }) => (
    <DropdownItem data-testid="search-context-menu-item" role="menuitem">
        <div className={classNames('d-flex', 'flex-column', styles.emptyResults)}>
            <span>
                <small>
                    <strong>{`Nobody found with the username “${userNameOrEmail}”`}</strong>
                </small>
            </span>
            <span className="text-muted">
                <small>Try sending invite via email instead</small>
            </span>
        </div>
    </DropdownItem>
)

const getUserSearchResultItem = (userId: string): HTMLButtonElement | null =>
    document.querySelector(`[data-res-user-id="${userId}"]`)

export const AutocompleteSearchUsers: React.FunctionComponent<
    React.PropsWithChildren<AutocompleteSearchUsersProps>
> = props => {
    const { disabled, onValueChanged, orgId } = props
    const MinSearchLength = 3
    const emailPattern = useRef(new RegExp(/^[\w!#$%&'*+./=?^`{|}~-]+@[A-Z_a-z]+?\.[A-Za-z]{2,3}$/))
    const [userNameOrEmail, setUsernameOrEmail] = useState('')
    const [isEmail, setIsEmail] = useState<boolean>(false)
    const inputReference = useRef<HTMLInputElement | null>(null)
    const [openResults, setOpenResults] = useState<boolean>(true)

    const [getUsers, { loading, data, error }] = useLazyQuery<
        AutocompleteMembersSearchResult,
        AutocompleteMembersSearchVariables
    >(SEARCH_USERS_AUTOCOMPLETE_QUERY, {
        variables: { organization: orgId, query: userNameOrEmail },
    })

    const results = (data
        ? data.autocompleteMembersSearch.map(usr => ({ ...usr })).sort(item => (item.inOrg ? 1 : -1))
        : []) as IUserItem[]

    const firstResult = results.length > 0 ? results[0] : undefined
    const resultsEnabled = !isEmail && !error && userNameOrEmail.length >= MinSearchLength && openResults
    const renderResults = resultsEnabled && results.length > 0
    const renderNoMatch = resultsEnabled && results.length === 0

    useEffect(() => {
        onValueChanged(userNameOrEmail, isEmail)
    }, [onValueChanged, isEmail, userNameOrEmail])

    const searchUsers = useCallback(
        (query: string): void => {
            setOpenResults(true)
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            getUsers({ variables: { query, organization: orgId } })
        },
        [getUsers, orgId]
    )

    const debounceGetUsers = useRef(debounce(searchUsers, 250, { leading: false }))

    const focusInputElement = (): void => {
        inputReference.current?.focus()
    }

    const onUsernameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        const newValue = event.currentTarget.value
        const isEmail = emailPattern.current.test(newValue)
        setIsEmail(isEmail)
        setUsernameOrEmail(newValue)
        if (!isEmail && newValue.length >= MinSearchLength) {
            debounceGetUsers.current(newValue)
        }
    }, [])

    const onInputKeyDown = useCallback(
        (event: React.KeyboardEvent) => {
            if (firstResult && event.key === 'ArrowDown') {
                event.stopPropagation()
                event.preventDefault()
                getUserSearchResultItem(firstResult.id)?.focus()
            } else if (event.key === 'Escape') {
                setOpenResults(false)
                event.stopPropagation()
            }
        },
        [firstResult]
    )

    const onSelectUser = useCallback(
        (user: IUserItem) => {
            eventLogger.log(
                'InviteAutocompleteUserSelected',
                { organizationId: orgId, user: user.username },
                { organizationId: orgId }
            )
            setOpenResults(false)
            setUsernameOrEmail(user.username)
        },
        [orgId]
    )

    const onMenuKeyDown = useCallback((event: React.KeyboardEvent): void => {
        if (event.key === 'Escape') {
            setOpenResults(false)
            event.stopPropagation()
            event.preventDefault()
            focusInputElement()
        }
    }, [])

    const toggleOpen = useCallback(() => {
        setOpenResults(false)
    }, [])

    const onResultItemKeydown = useCallback((key: string, index: number) => {
        if (index === 0 && key === 'ArrowUp') {
            window.requestAnimationFrame(() => focusInputElement())
        }
    }, [])

    return (
        <div className={styles.inputContainer} onKeyDown={onMenuKeyDown} tabIndex={-1} role="menu">
            <Dropdown isOpen={resultsEnabled} toggle={toggleOpen}>
                <DropdownToggle tag="div" data-toggle="dropdown">
                    <FormInput
                        autoFocus={true}
                        ref={inputReference}
                        value={userNameOrEmail}
                        label="Email address or username"
                        title="Email address or username"
                        onChange={onUsernameChange}
                        onKeyDown={onInputKeyDown}
                        aria-expanded={renderResults ? 'true' : 'false'}
                        disabled={disabled}
                        status={loading ? 'loading' : error ? 'error' : undefined}
                    />
                </DropdownToggle>
                <DropdownMenu className={styles.suggestionsContainer}>
                    {renderResults &&
                        results.map((usr, index) => (
                            <UserResultItem
                                key={usr.id}
                                index={index}
                                user={usr}
                                onSelectUser={onSelectUser}
                                onKeyDown={onResultItemKeydown}
                            />
                        ))}
                    {renderNoMatch && <EmptyResultsItem userNameOrEmail={userNameOrEmail} />}
                </DropdownMenu>
            </Dropdown>
        </div>
    )
}
