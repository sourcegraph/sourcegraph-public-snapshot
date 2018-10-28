import GlobeIcon from 'mdi-react/GlobeIcon'
import { RepoHeaderActionButton } from '../../../src/repo/RepoHeader'
import { repoHeaderActionButtons } from '../../../src/repo/repoHeaderActionButtons'
import { encodeRepoRev } from '../../../src/util/url'

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
