import { withKnobs } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import {
    RepositoryIndexConfigurationPage,
    RepositoryIndexConfigurationPageProps,
} from './RepositoryIndexConfigurationPage'

const story: Meta = {
    title: 'web/codeintel/configuration/RepositoryIndexConfigurationPage',
    decorators: [story => <div className="p-3 container">{story()}</div>, withKnobs],
    parameters: {
        component: RepositoryIndexConfigurationPage,
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}
export default story

const Template: Story<RepositoryIndexConfigurationPageProps> = args => (
    <WebStory mocks={[]}>{props => <RepositoryIndexConfigurationPage {...props} {...args} />}</WebStory>
)

const defaults: Partial<RepositoryIndexConfigurationPageProps> = {}

export const RepositoryPage = Template.bind({})
RepositoryPage.args = {
    ...defaults,
    repo: { id: '42' },
}
