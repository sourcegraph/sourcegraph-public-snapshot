import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'
import { RepositoryFields } from '../../graphql-operations'

import { RepositoryCommitsPage, RepositoryCommitsPageProps } from './RepositoryCommitsPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/RepositoryCommitsPage',
    decorators: [decorator],
}

export default config

const repo: RepositoryFields = {
    id: 'repo-id',
    name: 'github.com/sourcegraph/sourcegraph',
    url: 'https://github.com/sourcegraph/sourcegraph/perforce',
    isPerforceDepot: false,
    description: '',
    viewerCanAdminister: false,
    isFork: false,
    externalURLs: [],
    externalRepository: {
        __typename: 'ExternalRepository',
        serviceType: '',
        serviceID: '',
    },
    defaultBranch: null,
    metadata: [],
}

export const RepositoryCommitsPageStory: Story<RepositoryCommitsPageProps> = () => (
    <WebStory>{props => <RepositoryCommitsPage revision={''} {...props} repo={repo} />}</WebStory>
)

RepositoryCommitsPageStory.storyName = 'Repository commits page'
