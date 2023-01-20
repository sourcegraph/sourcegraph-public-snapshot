import { Meta, Story } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { ReferencesPanel } from './ReferencesPanel'
import { buildReferencePanelMocks, defaultProps } from './ReferencesPanel.mocks'

import webStyles from '../SourcegraphWebApp.scss'

const config: Meta = {
    title: 'wildcard/ReferencesPanel',
    component: ReferencesPanel,

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: ReferencesPanel,
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default config

export const Simple: Story = () => {
    const { url, requestMocks } = buildReferencePanelMocks()

    return (
        <div style={{ width: 1200, height: 400 }}>
            <MockedTestProvider mocks={requestMocks}>
                <ReferencesPanel {...defaultProps} initialActiveURL={url} />
            </MockedTestProvider>
        </div>
    )
}
