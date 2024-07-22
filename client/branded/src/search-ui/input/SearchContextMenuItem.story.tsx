import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { Combobox } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { SearchContextMenuItem } from './SearchContextMenu'

const decorator: Decorator = story => (
    <div className="dropdown-menu show" style={{ position: 'static' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'branded/search-ui/input/SearchContextMenuItem',
    parameters: {},
    decorators: [decorator],
}

export default config

export const SelectedDefaultItem: StoryFn = () => (
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

export const StarredItem: StoryFn = () => (
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
