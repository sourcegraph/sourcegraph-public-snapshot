import React, { FC, useCallback, useEffect, useRef, useState } from 'react'

import { useLazyQuery } from '@apollo/client'
import classNames from 'classnames'
import { debounce } from 'lodash'

import { Tooltip, Combobox, ComboboxInput, ComboboxPopover, ComboboxList, ComboboxOption } from '@sourcegraph/wildcard'

import { AutocompleteMembersSearchResult, AutocompleteMembersSearchVariables, Maybe } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'

import { SEARCH_USERS_AUTOCOMPLETE_QUERY } from './gqlQueries'

import styles from './SearchUserAutocomplete.module.scss'

const MIN_SEARCH_LENGTH = 3
const EMAIL_PATTERN = new RegExp(/^[\w!#$%&'*+./=?^`{|}~-]+@[A-Z_a-z]+?\.[A-Za-z]{2,3}$/)

interface AutocompleteSearchUsersProps {
    disabled?: boolean
    onValueChanged: (value: string, isEmail: boolean) => void
    orgId: string
}

export const AutocompleteSearchUsers: FC<AutocompleteSearchUsersProps> = props => {
    const { disabled, onValueChanged, orgId } = props

    const [userNameOrEmail, setUsernameOrEmail] = useState('')

    const [getUsers, { loading, data, error }] = useLazyQuery<
        AutocompleteMembersSearchResult,
        AutocompleteMembersSearchVariables
    >(SEARCH_USERS_AUTOCOMPLETE_QUERY, {
        variables: { organization: orgId, query: userNameOrEmail },
    })

    useEffect(() => {
        onValueChanged(userNameOrEmail, EMAIL_PATTERN.test(userNameOrEmail))
    }, [onValueChanged, userNameOrEmail])

    const searchUsers = useCallback(
        (query: string): void => {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            getUsers({ variables: { query, organization: orgId } })
        },
        [getUsers, orgId]
    )

    const debounceGetUsers = useRef(debounce(searchUsers, 250, { leading: false }))

    const handleUsernameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        const newValue = event.currentTarget.value

        setUsernameOrEmail(newValue)

        if (!EMAIL_PATTERN.test(newValue) && newValue.length >= MIN_SEARCH_LENGTH) {
            debounceGetUsers.current(newValue)
        }
    }, [])

    const handleUserSelect = (username: string): void => {
        setUsernameOrEmail(username)
        eventLogger.log(
            'InviteAutocompleteUserSelected',
            { organizationId: orgId, user: username },
            { organizationId: orgId }
        )
    }

    const results = (
        data ? data.autocompleteMembersSearch.map(usr => ({ ...usr })).sort(item => (item.inOrg ? 1 : -1)) : []
    ) as IUserItem[]

    const resultsEnabled = !EMAIL_PATTERN.test(userNameOrEmail) && !error && userNameOrEmail.length >= MIN_SEARCH_LENGTH
    const renderResults = resultsEnabled && results.length > 0
    const renderNoMatch = resultsEnabled && !loading && results.length === 0

    return (
        <Combobox className={styles.inputContainer} onSelect={handleUserSelect}>
            <ComboboxInput
                autocomplete={false}
                label="Email address or username"
                title="Email address or username"
                autoFocus={true}
                disabled={disabled}
                status={loading ? 'loading' : error ? 'error' : undefined}
                onChange={handleUsernameChange}
            />

            <ComboboxPopover className={styles.suggestionsContainer}>
                <ComboboxList>
                    {renderResults && results.map(usr => <UserResultItem key={usr.id} user={usr} />)}
                    {renderNoMatch && <EmptyResultsItem userNameOrEmail={userNameOrEmail} />}
                </ComboboxList>
            </ComboboxPopover>
        </Combobox>
    )
}

interface IUserItem {
    id: string
    username: string
    inOrg: boolean
    displayName: Maybe<string>
    avatarURL: Maybe<string>
}

interface UserResultItemProps {
    user: IUserItem
}

const UserResultItem: FC<UserResultItemProps> = props => {
    const { user } = props

    return (
        <ComboboxOption value={user.username} disabled={user.inOrg} data-res-user-id={user.id} className={styles.item}>
            <div className={styles.userContainer}>
                <div className={styles.avatarContainer}>
                    <Tooltip content={user.displayName || user.username}>
                        <UserAvatar
                            user={user}
                            size={24}
                            className={classNames(styles.avatar, user.inOrg && styles.avatarDisabled)}
                        />
                    </Tooltip>
                </div>
                <div className={styles.itemDescription}>
                    <div>
                        <strong>{user.displayName || user.username}</strong>{' '}
                        {user.displayName && <span className={styles.userName}>{user.username}</span>}
                    </div>
                    {user.inOrg && <small className={styles.userDescription}>Already in this organization</small>}
                </div>
            </div>
        </ComboboxOption>
    )
}

interface EmptyResultsItemProps {
    userNameOrEmail: string
}

const EmptyResultsItem: FC<EmptyResultsItemProps> = ({ userNameOrEmail }) => (
    <div role="menuitem" className={styles.emptyResults}>
        <span>
            <small>
                <strong>{`Nobody found with the username “${userNameOrEmail}”`}</strong>
            </small>
        </span>
        <span className="text-muted">
            <small>Try sending invite via email instead</small>
        </span>
    </div>
)
