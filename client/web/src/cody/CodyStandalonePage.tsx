import { useState } from 'react'

import { invoke } from '@tauri-apps/api/tauri'

import { gql, useQuery } from '@sourcegraph/http-client'
import { Label, Select } from '@sourcegraph/wildcard'

import { GetReposForCodyResult, GetReposForCodyVariables } from '../graphql-operations'

import { CodySidebar } from './sidebar/CodySidebar'
import { useChatStore } from './stores/chat'

interface CodyStandalonePageProps {}

const noop = (): void => {}

const REPOS_QUERY = gql`
    query GetReposForCody {
        repositories(first: 1000) {
            nodes {
                name
                embeddingExists
            }
        }
    }
`

export const CodyStandalonePage: React.FunctionComponent<CodyStandalonePageProps> = () => {
    const { data } = useQuery<GetReposForCodyResult, GetReposForCodyVariables>(REPOS_QUERY, {})

    const [selectedRepo, setSelectedRepo] = useState('github.com/sourcegraph/sourcegraph')
    useChatStore({ codebase: selectedRepo, setIsCodySidebarOpen: noop })

    const repos = data?.repositories.nodes ?? []

    return (
        <div className="d-flex flex-column w-100">
            <Label className="d-inline-flex align-items-center justify-content-center my-2 px-2 w-100">
                <span className="mr-2">Repo:</span>
                <Select
                    isCustomStyle={true}
                    className="mb-0"
                    aria-label="Select a repo"
                    id="repo-select"
                    value={selectedRepo || 'none'}
                    onChange={(event: React.ChangeEvent<HTMLSelectElement>): void => {
                        setSelectedRepo(event.target.value)
                    }}
                >
                    {repos.map(({ name }) => (
                        <option key={name} value={name}>
                            {name}
                        </option>
                    ))}
                </Select>
            </Label>

            <CodySidebar onClose={() => invoke('hide_window')} />
        </div>
    )
}
