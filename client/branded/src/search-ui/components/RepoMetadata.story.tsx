import { Meta, Story } from '@storybook/react'

import { Text } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { RepoMetadata } from './RepoMetadata'

const config: Meta = {
    title: 'branded/search-ui/RepoMetadata',
    parameters: {
        chromatic: { viewports: [480] },
    },
}

export default config

const mockMetadata: [string, string | undefined][] = [
    ['team', 'iam'],
    ['org', 'source'],
    ['oss', undefined],
]

export const RepoMetadataStory: Story = () => (
    <BrandedStory>
        {() => (
            <div className="m-3">
                <div className="d-flex align-items-center mb-2">
                    <Text className="mb-0 mr-3 text-no-wrap">Default</Text>
                    <RepoMetadata metadata={mockMetadata} />
                </div>
                <div className="d-flex align-items-center mb-2">
                    <Text className="mb-0 mr-3 text-no-wrap">Default & Delete mode</Text>
                    <RepoMetadata keyValuePairs={mockMetadata} onDelete={key => alert(key)} />
                </div>
                <div className="d-flex align-items-center mb-2">
                    <Text className="mb-0 mr-3 text-no-wrap">Small</Text>
                    <RepoMetadata metadata={mockMetadata} small={true} />
                </div>
                <div className="d-flex align-items-center mb-2">
                    <Text className="mb-0 mr-3 text-no-wrap">Small & Delete mode</Text>
                    <RepoMetadata keyValuePairs={mockMetadata} small={true} onDelete={key => alert(key)} />
                </div>
            </div>
        )}
    </BrandedStory>
)

RepoMetadataStory.storyName = 'RepoMetadata'
RepoMetadataStory.parameters = {
    chromatic: {
        disableSnapshot: false,
    },
}
