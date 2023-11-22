import { type FC, useState } from 'react'

import { gql, useQuery } from '@apollo/client'
import { mdiGithub } from '@mdi/js'
import { identity } from 'lodash'

import {
    ErrorAlert,
    Icon,
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxList,
    MultiComboboxOption,
    MultiComboboxOptionText,
    useDebounce,
} from '@sourcegraph/wildcard'

import type {
    GetGitHubOrganizationsResult,
    GetGitHubOrganizationsVariables,
    GetGitHubRepositoriesResult,
    GetGitHubRepositoriesVariables,
} from '../../../../../../graphql-operations'

import styles from './GithubEntityPickers.module.scss'

const GET_GITHUB_ORGANIZATIONS = gql`
    query GetGitHubOrganizations($id: ID, $token: String!) {
        externalServiceNamespaces(kind: GITHUB, url: "https://github.com", token: $token, id: $id) {
            nodes {
                id
                name
            }
        }
    }
`

interface GithubOrganizationsPickerProps {
    token: string
    disabled: boolean
    organizations: string[]
    externalServiceId: string | undefined
    onChange: (orginaziations: string[]) => void
}

export const GithubOrganizationsPicker: FC<GithubOrganizationsPickerProps> = props => {
    const { token, disabled, organizations, externalServiceId, onChange } = props
    const [searchTerm, setSearchTerm] = useState('')

    const { data, loading, error } = useQuery<GetGitHubOrganizationsResult, GetGitHubOrganizationsVariables>(
        GET_GITHUB_ORGANIZATIONS,
        {
            skip: disabled,
            variables: { token, id: externalServiceId ?? null },
        }
    )

    const handleSelectedItemsChange = (orginaziations: string[]): void => {
        setSearchTerm('')
        onChange(orginaziations)
    }

    const suggestions = (data?.externalServiceNamespaces?.nodes ?? []).map(item => item.name)

    // Render only non-selected organizations and orgs that match search term value
    const filteredSuggestions = suggestions.filter(
        orgName => !organizations.includes(orgName) && orgName.toLowerCase().includes(searchTerm.toLowerCase())
    )

    return (
        <MultiCombobox
            selectedItems={organizations}
            getItemKey={identity}
            getItemName={identity}
            onSelectedItemsChange={handleSelectedItemsChange}
        >
            <MultiComboboxInput
                value={searchTerm}
                disabled={disabled}
                placeholder="Search organization"
                status={loading ? 'loading' : 'initial'}
                onChange={event => setSearchTerm(event.target.value)}
            />
            <small className="d-block text-muted pl-2 mt-2">
                Pick at least one organization and we clone all repositories that this organization has
            </small>

            <MultiComboboxList items={filteredSuggestions} className="mt-2">
                {items =>
                    items.map((item, index) => (
                        <MultiComboboxOption key={item} value={item} index={index} className={styles.item}>
                            <Icon aria-hidden={true} svgPath={mdiGithub} /> <MultiComboboxOptionText />
                        </MultiComboboxOption>
                    ))
                }
            </MultiComboboxList>

            {error && <ErrorAlert error={error} className="mt-3" />}
        </MultiCombobox>
    )
}

const GET_GITHUB_REPOSITORIES = gql`
    query GetGitHubRepositories(
        $id: ID
        $first: Int!
        $token: String!
        $query: String!
        $excludeRepositories: [String!]!
    ) {
        externalServiceRepositories(
            kind: GITHUB
            url: "https://github.com"
            id: $id
            first: $first
            token: $token
            query: $query
            excludeRepos: $excludeRepositories
        ) {
            nodes {
                id
                name
            }
        }
    }
`

interface GithubRepositoriesPickerProps {
    token: string
    disabled: boolean
    repositories: string[]
    externalServiceId: string | undefined
    onChange: (repositories: string[]) => void
}

export const GithubRepositoriesPicker: FC<GithubRepositoriesPickerProps> = props => {
    const { token, disabled, repositories, externalServiceId, onChange } = props

    const [searchTerm, setSearchTerm] = useState('')

    const {
        data: currentData,
        previousData,
        loading,
        error,
    } = useQuery<GetGitHubRepositoriesResult, GetGitHubRepositoriesVariables>(GET_GITHUB_REPOSITORIES, {
        skip: disabled,
        fetchPolicy: 'cache-and-network',
        variables: {
            token,
            first: 10,
            query: useDebounce(searchTerm, 500),
            id: externalServiceId ?? null,
            excludeRepositories: repositories,
        },
    })

    const handleSelectedItemsChange = (repositories: string[]): void => {
        setSearchTerm('')
        onChange(repositories)
    }

    const data = currentData ?? previousData
    const suggestions = (data?.externalServiceRepositories?.nodes ?? []).map(item => formatRepositoryName(item.name))

    // Render only non-selected repositories and repositories that match search term value
    const filteredSuggestions = suggestions.filter(
        repoName => !repositories.includes(repoName) && repoName.toLowerCase().includes(searchTerm.toLowerCase())
    )

    return (
        <MultiCombobox
            selectedItems={repositories}
            getItemKey={identity}
            getItemName={identity}
            onSelectedItemsChange={handleSelectedItemsChange}
        >
            <MultiComboboxInput
                value={searchTerm}
                placeholder="Search repository"
                disabled={disabled}
                status={loading ? 'loading' : 'initial'}
                onChange={event => setSearchTerm(event.target.value)}
            />
            <small className="d-block text-muted pl-2 mt-2">Pick at least one repository</small>

            <MultiComboboxList
                renderEmptyList={true}
                items={filteredSuggestions}
                className={styles.repositoriesSuggest}
            >
                {items =>
                    items.map((item, index) => (
                        <MultiComboboxOption key={item} value={item} index={index} className={styles.item}>
                            <Icon aria-hidden={true} svgPath={mdiGithub} /> <MultiComboboxOptionText />
                        </MultiComboboxOption>
                    ))
                }
            </MultiComboboxList>

            {error && <ErrorAlert error={error} className="mt-3" />}
        </MultiCombobox>
    )
}

/**
 * Format GitHub repository URLs and return repository name (URL withouth code host prefix)
 * For example: "github.com/sourcegraph/about" will become "sourcegraph/about"
 */
function formatRepositoryName(url: string): string {
    return url.slice(Math.max(0, url.indexOf('/') + 1))
}
