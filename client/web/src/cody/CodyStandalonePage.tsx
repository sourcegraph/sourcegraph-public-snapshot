import { gql, useQuery } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Select } from '@sourcegraph/wildcard'

import { GetReposForCodyResult, GetReposForCodyVariables } from '../graphql-operations'

import { CodySidebar } from './sidebar/CodySidebar'
import { useChatStore } from './stores/chat'

import styles from './CodyStandalonePage.module.scss'

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

export const CodyStandalonePage: React.FunctionComponent<{}> = () => {
    const { data } = useQuery<GetReposForCodyResult, GetReposForCodyVariables>(REPOS_QUERY, {})

    const [selectedRepo, setSelectedRepo] = useTemporarySetting('app.codyStandalonePage.selectedRepo', '')
    useChatStore({ codebase: selectedRepo || '', setIsCodySidebarOpen: noop })

    const repos = data?.repositories.nodes ?? []

    const repoSelector = (
        <Select
            isCustomStyle={true}
            className={styles.repoSelect}
            selectSize="sm"
            label="Repo:"
            id="repo-select"
            value={selectedRepo || 'none'}
            onChange={(event: React.ChangeEvent<HTMLSelectElement>): void => {
                setSelectedRepo(event.target.value)
            }}
        >
            <option value="" disabled={true}>
                Select a repo
            </option>
            {repos.map(({ name }) => (
                <option key={name} value={name}>
                    {name}
                </option>
            ))}
        </Select>
    )

    return (
        <div className="d-flex flex-column w-100">
            <CodySidebar titleContent={repoSelector} />
        </div>
    )
}
