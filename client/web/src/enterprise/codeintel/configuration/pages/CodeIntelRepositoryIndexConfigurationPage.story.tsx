import { withKnobs } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import {
    CodeIntelRepositoryIndexConfigurationPage,
    CodeIntelRepositoryIndexConfigurationPageProps,
} from './CodeIntelRepositoryIndexConfigurationPage'

const story: Meta = {
    title: 'web/codeintel/configuration/CodeIntelRepositoryIndexConfigurationPage',
    decorators: [story => <div className="p-3 container">{story()}</div>, withKnobs],
    parameters: {
        component: CodeIntelRepositoryIndexConfigurationPage,
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}
export default story

const Template: Story<CodeIntelRepositoryIndexConfigurationPageProps> = args => (
    <WebStory mocks={[]}>{props => <CodeIntelRepositoryIndexConfigurationPage {...props} {...args} />}</WebStory>
)

const defaults: Partial<CodeIntelRepositoryIndexConfigurationPageProps> = {}

export const RepositoryPage = Template.bind({})
RepositoryPage.args = {
    ...defaults,
    repo: { id: '42' },
}
RepositoryPage.parameters = {
    // Keep snapshots for one variant
    chromatic: { disableSnapshots: false },
}
