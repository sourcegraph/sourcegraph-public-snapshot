import { useLazyQuery } from '@apollo/client'
import classNames from 'classnames'
import { debounce } from 'lodash'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { DropdownItem } from 'reactstrap'
// import { Key } from 'ts-key-enum'

import { Input } from '@sourcegraph/wildcard'

import { AutocompleteUsersResult, AutocompleteUsersVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { SEARCH_USERS_AUTOCOMPLETE_QUERY } from './gqlQueries'
import styles from './SearchUserAutocomplete.module.scss'

interface IUserItem {
    id: string
    username: string
    displayName?: string
    avatarUrl?: string
}

interface AutocompleteSearchUsersProps {
    disabled?: boolean
    onValueChanged: (value: string, isEmail: boolean) => void
}

export const UserResultItem: React.FunctionComponent<{
    onKeyDown: (key: string) => void
    onSelectUser: (user: IUserItem) => void
    user: IUserItem
}> = ({ onKeyDown, user, onSelectUser }) => {
    const selectUser = useCallback(() => {
        onSelectUser(user)
    }, [onSelectUser, user])

    const itemKeyDown = useCallback(
        (event: React.KeyboardEvent) => {
            onKeyDown(event.key)
            event.stopPropagation()
            event.preventDefault()
        },
        [onKeyDown]
    )

    return (
        <DropdownItem
            data-testid="search-context-menu-item"
            data-res-user-id={user.id}
            toggle={false}
            className={styles.item}
            onClick={selectUser}
            role="menuitem"
            onKeyDown={itemKeyDown}
        >
            <div>
                {user.username} - {user.displayName} - {user.avatarUrl}
            </div>
        </DropdownItem>
    )
}

const getUserSearchResultItem = (userId: string): HTMLButtonElement | null =>
    document.querySelector(`[data-res-user-id="${userId}"]`)

export const AutocompleteSearchUsers: React.FunctionComponent<AutocompleteSearchUsersProps> = props => {
    const { disabled, onValueChanged } = props
    const MinSearchLength = 3
    const emailPattern = useRef(new RegExp(/^[\w!#$%&'*+./=?^`{|}~-]+@[A-Z_a-z]+?\.[A-Za-z]{2,3}$/))
    const [userNameOrEmail, setUsernameOrEmail] = useState('')
    const [isEmail, setIsEmail] = useState<boolean>(false)
    const resultList = useRef<HTMLDivElement | null>(null)
    const inputReference = useRef<HTMLInputElement | null>(null)
    const [openResults, setOpenResults] = useState<boolean>(true)

    const [getUsers, { loading, data, error }] = useLazyQuery<AutocompleteUsersResult, AutocompleteUsersVariables>(
        SEARCH_USERS_AUTOCOMPLETE_QUERY,
        {
            variables: { query: userNameOrEmail },
        }
    )

    const results = (data ? data.autocompleteSearchUsers || [] : []) as IUserItem[]
    const firstResult = results.length > 0 ? results[0] : undefined
    const resultsEnabled = !isEmail && !loading && !error && userNameOrEmail.length >= MinSearchLength && openResults
    const renderResults = resultsEnabled && results.length > 0
    const renderNoMatch = resultsEnabled && results.length === 0

    useEffect(() => {
        onValueChanged(userNameOrEmail, isEmail)
    }, [onValueChanged, isEmail, userNameOrEmail])

    const searchUsers = useCallback(
        (query: string): void => {
            setOpenResults(true)
            getUsers({ variables: { query } })
        },
        [getUsers]
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

    const onSelectUser = useCallback((user: IUserItem) => {
        eventLogger.logViewEvent('AutocompleteUsersSearchSelected', { user: user.username })
        setOpenResults(false)
        setUsernameOrEmail(user.username)
    }, [])

    const onMenuKeyDown = useCallback((key: string, index: number): void => {
        if (key === 'Escape') {
            setOpenResults(false)
        } else if (key === 'ArrowUp' && index === 0) {
            focusInputElement()
        }
    }, [])

    return (
        <div className={styles.inputContainer}>
            <Input
                autoFocus={true}
                ref={inputReference}
                value={userNameOrEmail}
                label="Email address or username"
                title="Email address or username"
                onChange={onUsernameChange}
                onKeyDown={onInputKeyDown}
                disabled={disabled}
                status={loading ? 'loading' : error ? 'error' : undefined}
            />
            <div
                data-testid="search-context-menu-list"
                className={styles.suggestionsContainer}
                ref={resultList}
                role="menu"
            >
                {renderResults &&
                    results.map((usr, index) => (
                        <UserResultItem
                            key={usr.id}
                            user={usr}
                            onSelectUser={onSelectUser}
                            onKeyDown={key => onMenuKeyDown(key, index)}
                        />
                    ))}
                {renderNoMatch && (
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
                )}
            </div>
        </div>
    )
}
