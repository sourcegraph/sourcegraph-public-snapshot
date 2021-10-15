import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { SearchContextMenuItem } from './SearchContextMenu'

const { add } = storiesOf('web/searchContexts/SearchContextMenuItem', module)
    .addParameters({
        chromatic: { viewports: [1200] },
    })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

add(
    'selected default item',
    () => (
        <BrandedStory>
            {() => (
                <SearchContextMenuItem
                    spec="@user/test"
                    searchFilter=""
                    description="Default description"
                    selected={true}
                    isDefault={true}
                    selectSearchContextSpec={noop}
                    onKeyDown={noop}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            )}
        </BrandedStory>
    ),
    {}
)

add(
    'highlighted item',
    () => (
        <BrandedStory>
            {() => (
                <SearchContextMenuItem
                    spec="@user/test"
                    searchFilter="@us/te"
                    description="Default description"
                    selected={false}
                    isDefault={false}
                    selectSearchContextSpec={noop}
                    onKeyDown={noop}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            )}
        </BrandedStory>
    ),
    {}
)
