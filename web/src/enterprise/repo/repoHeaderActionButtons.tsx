import { GlobeIcon } from 'mdi-react'
import { RepoHeaderActionButton } from '../../repo/RepoHeader'
import { repoHeaderActionButtons } from '../../repo/repoHeaderActionButtons'
import { encodeRepoRev } from '../../util/url'

const enableRepositoryGraph = localStorage.getItem('repositoryGraph') !== null

export const enterpriseRepoHeaderActionButtons: ReadonlyArray<RepoHeaderActionButton> = [
    ...repoHeaderActionButtons,
    {
        label: 'Graph',
        icon: GlobeIcon,
        condition: () => enableRepositoryGraph,
        to: context => `/${encodeRepoRev(context.repoName, context.encodedRev)}/-/graph`,
        tooltip: 'Repository graph',
    },
]
