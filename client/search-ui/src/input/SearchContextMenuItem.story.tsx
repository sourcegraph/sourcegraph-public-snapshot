import { Meta, Story, DecoratorFn } from '@storybook/react'
import { noop } from 'lodash'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { SearchContextMenuItem } from './SearchContextMenu'

const decorator: DecoratorFn = story => (
    <div className="dropdown-menu show" style={{ position: 'static' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'search-ui/input/SearchContextMenuItem',
    parameters: {
        chromatic: { viewports: [1200], disableSnapshot: false },
    },
    decorators: [decorator],
}

export default config

export const SelectedDefaultItem: Story = () => (
    <BrandedStory>
        {() => (
            <SearchContextMenuItem
                spec="@user/test"
                searchFilter=""
                description="Default description"
                query=""
                selected={true}
                isDefault={true}
                selectSearchContextSpec={noop}
                onKeyDown={noop}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </BrandedStory>
)

SelectedDefaultItem.storyName = 'selected default item'

export const HighlightedItem: Story = () => (
    <BrandedStory>
        {() => (
            <SearchContextMenuItem
                spec="@user/test"
                searchFilter="@us/te"
                description="Default description"
                query=""
                selected={false}
                isDefault={false}
                selectSearchContextSpec={noop}
                onKeyDown={noop}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </BrandedStory>
)

HighlightedItem.storyName = 'highlighted item'
