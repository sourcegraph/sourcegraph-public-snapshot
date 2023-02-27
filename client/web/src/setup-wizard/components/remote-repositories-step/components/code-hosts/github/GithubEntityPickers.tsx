import { FC, useState } from 'react'

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
} from '@sourcegraph/wildcard'

import { GetGitHubOrganizationsResult, GetGitHubOrganizationsVariables } from '../../../../../../graphql-operations'

const GET_GITHUB_ORGANIZATIONS = gql`
    query GetGitHubOrganizations($token: String!) {
        externalServiceNamespaces(kind: GITHUB, url: "https://github.com", token: $token) {
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
    onChange: (orginaziations: string[]) => void
}

export const GithubOrganizationsPicker: FC<GithubOrganizationsPickerProps> = props => {
    const { token, disabled, organizations, onChange } = props
    const [searchTerm, setSearchTerm] = useState('')

    const { data, loading, error } = useQuery<GetGitHubOrganizationsResult, GetGitHubOrganizationsVariables>(
        GET_GITHUB_ORGANIZATIONS,
        {
            skip: disabled,
            variables: { token },
        }
    )

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
            onSelectedItemsChange={onChange}
        >
            <MultiComboboxInput
                value={searchTerm}
                disabled={disabled}
                placeholder="Search orgnaization"
                status={loading ? 'loading' : 'initial'}
                onChange={event => setSearchTerm(event.target.value)}
            />
            <small className="text-muted pl-2">
                Pick at least one organization and we clone all repositories that this organzization has
            </small>

            <MultiComboboxList items={filteredSuggestions} className="mt-2">
                {items =>
                    items.map((item, index) => (
                        <MultiComboboxOption key={item} value={item} index={index}>
                            <Icon aria-hidden={true} svgPath={mdiGithub} /> <MultiComboboxOptionText />
                        </MultiComboboxOption>
                    ))
                }
            </MultiComboboxList>

            {error && <ErrorAlert error={error} className="mt-3" />}
        </MultiCombobox>
    )
}

const DEMO_REPOS_SUGGESTIONS = [
    'sourcegraph/sourcegraph',
    'sourcegraph/about',
    'personal/my-project',
    'peraonal/university-labs',
    'facebook/react',
]

interface GithubRepositoriesPickerProps {
    repositories: string[]
    onChange: (repositories: string[]) => void
}

export const GithubRepositoriesPicker: FC<GithubRepositoriesPickerProps> = props => {
    const { repositories, onChange } = props

    const [searchTerm, setSearchTerm] = useState('')

    const suggestions = DEMO_REPOS_SUGGESTIONS.filter(item => !repositories.find(selectedItem => selectedItem === item))

    return (
        <MultiCombobox
            selectedItems={repositories}
            getItemKey={identity}
            getItemName={identity}
            onSelectedItemsChange={onChange}
        >
            <MultiComboboxInput
                value={searchTerm}
                placeholder="Search repository"
                onChange={event => setSearchTerm(event.target.value)}
            />
            <small className="text-muted pl-2">Pick at least one repository</small>

            <MultiComboboxList items={suggestions} className="mt-2">
                {items =>
                    items.map((item, index) => (
                        <MultiComboboxOption key={item} value={item} index={index}>
                            <Icon aria-hidden={true} svgPath={mdiGithub} /> <MultiComboboxOptionText />
                        </MultiComboboxOption>
                    ))
                }
            </MultiComboboxList>
        </MultiCombobox>
    )
}
