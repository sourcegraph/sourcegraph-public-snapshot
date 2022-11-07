import { Meta, Story } from '@storybook/react'
import * as H from 'history'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { ReferencesPanelWithMemoryRouter } from './ReferencesPanel'
import { buildReferencePanelMocks, defaultProps } from './ReferencesPanel.mocks'

const config: Meta = {
    title: 'wildcard/ReferencesPanel',
    component: ReferencesPanelWithMemoryRouter,

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: ReferencesPanelWithMemoryRouter,
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default config

export const Simple: Story = () => {
    const { url, requestMocks } = buildReferencePanelMocks()

    const fakeExternalHistory = H.createMemoryHistory()
    fakeExternalHistory.push(url)

    return (
        <div style={{ width: 1200, height: 400 }}>
            <MockedTestProvider mocks={requestMocks}>
                <ReferencesPanelWithMemoryRouter
                    {...defaultProps}
                    externalHistory={fakeExternalHistory}
                    externalLocation={fakeExternalHistory.location}
                    initialActiveURL="/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/diff/diff.go?L16:2-16:10"
                />
            </MockedTestProvider>
        </div>
    )
}
