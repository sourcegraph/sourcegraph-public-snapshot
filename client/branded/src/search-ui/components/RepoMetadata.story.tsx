import { Meta, Story } from '@storybook/react'

import { Card, Text } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { RepoMetadataItem, RepoMetadata } from './RepoMetadata'

const config: Meta = {
    title: 'branded/search-ui/RepoMetadata',
    parameters: {
        chromatic: { viewports: [480] },
    },
}

export default config

const mockItems: RepoMetadataItem[] = [
    {
        key: 'archived',
        value: 'true',
    },
    {
        key: 'oss',
    },
    {
        key: 'license',
        value: 'multiple',
    },
]

export const RepoMetadataStory: Story = () => (
    <BrandedStory>
        {() => (
            <Card className="p-3">
                <div className="d-flex align-items-center mb-3">
                    <Text className="mb-0 mr-3 text-no-wrap">Default</Text>
                    <RepoMetadata items={mockItems} />
                </div>
                <div className="d-flex align-items-center mb-3">
                    <Text className="mb-0 mr-3 text-no-wrap">Deletable metadata</Text>
                    <RepoMetadata items={mockItems} onDelete={key => alert(key)} />
                </div>
                <div className="d-flex align-items-center mb-3">
                    <Text className="mb-0 mr-3 text-no-wrap">Small</Text>
                    <RepoMetadata items={mockItems} small={true} />
                </div>
                <div className="d-flex align-items-center">
                    <Text className="mb-0 mr-3 text-no-wrap">Small & Deletable metadata</Text>
                    <RepoMetadata items={mockItems} onDelete={key => alert(key)} small={true} />
                </div>
            </Card>
        )}
    </BrandedStory>
)

RepoMetadataStory.storyName = 'RepoMetadata'
RepoMetadataStory.parameters = {
    chromatic: {
        disableSnapshot: false,
    },
}
