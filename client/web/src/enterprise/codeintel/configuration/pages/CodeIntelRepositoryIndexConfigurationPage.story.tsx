import { withKnobs } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'

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
