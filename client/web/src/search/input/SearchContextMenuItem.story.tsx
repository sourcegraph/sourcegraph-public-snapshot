import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../components/WebStory'

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
        <WebStory>
            {() => (
                <SearchContextMenuItem
                    spec="@user/test"
                    searchFilter=""
                    description="Default description"
                    selected={true}
                    isDefault={true}
                    selectSearchContextSpec={() => {}}
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'highlighted item',
    () => (
        <WebStory>
            {() => (
                <SearchContextMenuItem
                    spec="@user/test"
                    searchFilter="@us/te"
                    description="Default description"
                    selected={false}
                    isDefault={false}
                    selectSearchContextSpec={() => {}}
                />
            )}
        </WebStory>
    ),
    {}
)
