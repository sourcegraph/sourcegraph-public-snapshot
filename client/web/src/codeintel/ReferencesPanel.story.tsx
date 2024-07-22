import type { Meta, StoryFn } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { ReferencesPanel } from './ReferencesPanel'
import { buildReferencePanelMocks, defaultProps } from './ReferencesPanel.mocks'

import webStyles from '../SourcegraphWebApp.scss'

const config: Meta = {
    title: 'wildcard/ReferencesPanel',
    component: ReferencesPanel,

    parameters: {
        component: ReferencesPanel,
    },
}

export default config

export const Simple: StoryFn = () => {
    const { url, requestMocks } = buildReferencePanelMocks()

    return (
        <BrandedStory styles={webStyles} initialEntries={[url]}>
            {() => (
                <div className="container mt-3 pb-3">
                    <div style={{ width: 1200, height: 400 }}>
                        <MockedTestProvider mocks={requestMocks}>
                            <ReferencesPanel {...defaultProps} initialActiveURL={url} />
                        </MockedTestProvider>
                    </div>
                </div>
            )}
        </BrandedStory>
    )
}
