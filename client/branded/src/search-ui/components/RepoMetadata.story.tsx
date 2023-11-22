import type { Meta, StoryFn } from '@storybook/react'

import { Card, Grid, H2, H3 } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { type RepoMetadataItem, RepoMetadata } from './RepoMetadata'

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

export const RepoMetadataStory: StoryFn = () => (
    <BrandedStory>
        {() => (
            <Card className="p-3">
                <Grid columnCount={3}>
                    <H3 className="mb-0 mr-3 text-no-wrap">Default</H3>
                    <H2 className="mb-0 mr-3 text-no-wrap">Link</H2>
                    <H3 className="mb-0 mr-3 text-no-wrap">Delete</H3>
                    <RepoMetadata items={mockItems} />
                    <RepoMetadata
                        items={mockItems}
                        queryState={{ query: '' }}
                        buildSearchURLQueryFromQueryState={() => ''}
                    />
                    <RepoMetadata items={mockItems} onDelete={key => alert(key)} />
                </Grid>
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
