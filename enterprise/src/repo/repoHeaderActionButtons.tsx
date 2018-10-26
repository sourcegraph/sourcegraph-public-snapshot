import { RepoHeaderActionButton } from '@sourcegraph/webapp/dist/repo/RepoHeader'
import { repoHeaderActionButtons } from '@sourcegraph/webapp/dist/repo/repoHeaderActionButtons'
import { encodeRepoRev } from '@sourcegraph/webapp/dist/util/url'
import GlobeIcon from 'mdi-react/GlobeIcon'

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
