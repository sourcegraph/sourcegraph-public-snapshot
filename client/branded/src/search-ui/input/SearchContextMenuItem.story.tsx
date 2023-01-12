import { Meta, Story, DecoratorFn } from '@storybook/react'

import { Combobox } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { SearchContextMenuItem } from './SearchContextMenu'

const decorator: DecoratorFn = story => (
    <div className="dropdown-menu show" style={{ position: 'static' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'branded/search-ui/input/SearchContextMenuItem',
    parameters: {
        chromatic: { viewports: [1200], disableSnapshot: false },
    },
    decorators: [decorator],
}

export default config

export const SelectedDefaultItem: Story = () => (
    <BrandedStory>
        {() => (
            <Combobox>
                <SearchContextMenuItem
                    spec="@user/test"
                    description="Default description"
                    query=""
                    selected={true}
                    isDefault={true}
                    starred={false}
                />
            </Combobox>
        )}
    </BrandedStory>
)

SelectedDefaultItem.storyName = 'selected default item'

export const StarredItem: Story = () => (
    <BrandedStory>
        {() => (
            <Combobox>
                <SearchContextMenuItem
                    spec="@user/test"
                    description="Default description"
                    query=""
                    selected={false}
                    isDefault={false}
                    starred={true}
                />
            </Combobox>
        )}
    </BrandedStory>
)

StarredItem.storyName = 'starred item'
