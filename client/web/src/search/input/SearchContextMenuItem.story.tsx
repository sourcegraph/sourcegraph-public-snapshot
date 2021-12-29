import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
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
                    query=""
                    selected={true}
                    isDefault={true}
                    selectSearchContextSpec={noop}
                    onKeyDown={noop}
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
                    query=""
                    selected={false}
                    isDefault={false}
                    selectSearchContextSpec={noop}
                    onKeyDown={noop}
                />
            )}
        </WebStory>
    ),
    {}
)
