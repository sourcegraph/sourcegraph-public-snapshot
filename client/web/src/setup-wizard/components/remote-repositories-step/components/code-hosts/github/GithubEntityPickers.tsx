import { FC, useState } from 'react'

import { mdiGithub } from '@mdi/js'

import {
    Icon,
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxList,
    MultiComboboxOption,
    MultiComboboxOptionText,
} from '@sourcegraph/wildcard'
import { identity } from 'lodash'

const DEMO_ORGS_SUGGESTIONS = [
    'Sourcegraph org',
    'My personal organization',
    'Golang working group',
    'University labs',
    'React working group',
]

interface GithubOrganizationsPickerProps {
    organizations: string[]
    onChange: (orginaziations: string[]) => void
}

export const GithubOrganizationsPicker: FC<GithubOrganizationsPickerProps> = props => {
    const { organizations, onChange } = props
    const [searchTerm, setSearchTerm] = useState('')

    const suggestions = DEMO_ORGS_SUGGESTIONS.filter(item => !organizations.find(selectedItem => selectedItem === item))

    return (
        <MultiCombobox
            selectedItems={organizations}
            getItemKey={identity}
            getItemName={identity}
            onSelectedItemsChange={onChange}
        >
            <MultiComboboxInput
                value={searchTerm}
                placeholder="Search orgnaization"
                onChange={event => setSearchTerm(event.target.value)}
            />
            <small className="text-muted pl-2">
                Pick at least one organization and we clone all repositories that this organzization has
            </small>

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
