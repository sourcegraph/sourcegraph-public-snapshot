import { type FC, useState } from 'react'

import { gql, useQuery } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import {
    ErrorAlert,
    LoadingSpinner,
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxList,
    MultiComboboxOption,
    MultiComboboxPopover,
    MultiComboboxOptionText,
    useDebounce,
} from '@sourcegraph/wildcard'

import type { GetUsersListResult, GetUsersListVariables } from '../../graphql-operations'

import styles from './UsersPicker.module.scss'

/**
 * User picker query to fetch list of users, exported only for Storybook story
 * apollo mocks, not designed to be reused in other places.
 */
export const GET_USERS_QUERY = gql`
    fragment UserSuggestion on User {
        id
        username
        displayName
        avatarURL
        siteAdmin
        primaryEmail {
            email
        }
    }

    query GetUsersList($query: String!) {
        users(first: 15, query: $query) {
            nodes {
                ...UserSuggestion
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }
`

export interface User {
    id: string
    username: string
    avatarURL: string | null
}

interface UsersPickerProps {
    value: User[]
    onChange: (users: User[]) => void
}

export const UsersPicker: FC<UsersPickerProps> = props => {
    const { value, onChange } = props

    const [searchTerm, setSearchTerm] = useState('')
    const debouncedSearchTerm = useDebounce(searchTerm, 500)

    const {
        data: currentData,
        previousData,
        loading,
        error,
    } = useQuery<GetUsersListResult, GetUsersListVariables>(GET_USERS_QUERY, {
        variables: {
            query: debouncedSearchTerm,
        },
        fetchPolicy: 'cache-and-network',
    })

    const data = currentData ?? previousData
    const suggestions = data !== undefined ? data.users.nodes : []
    const selectedUsersIds = new Set(value.map(user => user.id))
    const filteredSuggestions = suggestions.filter(user => !selectedUsersIds.has(user.id))

    const hasNextPage = data?.users.pageInfo.hasNextPage ?? false

    return (
        <MultiCombobox
            selectedItems={value}
            getItemKey={user => user.id}
            getItemName={user => user.username}
            onSelectedItemsChange={onChange}
        >
            <MultiComboboxInput
                value={searchTerm}
                status={loading ? 'loading' : 'initial'}
                placeholder="Filter by users..."
                autoCorrect="false"
                autoComplete="off"
                spellCheck={false}
                getPillContent={user => <CustomUserPickerPill user={user} />}
                onChange={event => setSearchTerm(event.target.value)}
            />

            <MultiComboboxPopover syncWidth={false} className={styles.popover}>
                {loading && filteredSuggestions.length === 0 && (
                    <>
                        <LoadingSpinner /> Fetching users...
                    </>
                )}

                {!loading && error && <ErrorAlert error={error} />}

                {filteredSuggestions.length > 0 && (
                    <MultiComboboxList items={filteredSuggestions}>
                        {users =>
                            users.map((user, index) => (
                                <MultiComboboxOption
                                    key={user.id}
                                    value={user.username}
                                    index={index}
                                    className={styles.item}
                                >
                                    <UserAvatar user={user} className={styles.itemAvatar} />
                                    <span className={styles.itemUsername}>
                                        <MultiComboboxOptionText />
                                    </span>
                                    {user.siteAdmin && <span className={styles.itemRole}>Admin</span>}

                                    <span className={styles.itemEmail}>
                                        {user.primaryEmail?.email ?? 'No email set'}
                                    </span>
                                </MultiComboboxOption>
                            ))
                        }
                    </MultiComboboxList>
                )}

                {hasNextPage && (
                    <footer className={styles.footer}>The first 15 matches are shown, narrow down you search</footer>
                )}
            </MultiComboboxPopover>
        </MultiCombobox>
    )
}

interface CustomUserPickerPillProps {
    user: User
}

const CustomUserPickerPill: FC<CustomUserPickerPillProps> = ({ user }) => (
    <span className={styles.pill}>
        <UserAvatar
            user={{
                avatarURL: user.avatarURL,
                username: user.username,
                displayName: null,
            }}
            className={styles.pillAvatar}
        />
        <span className={styles.pillText}>{user.username}</span>
    </span>
)
