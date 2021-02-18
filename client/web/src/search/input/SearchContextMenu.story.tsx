import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { SearchContextMenu, SearchContextMenuProps } from './SearchContextMenu'

const { add } = storiesOf('web/search/input/SearchContextMenu', module)
    .addParameters({
        chromatic: { viewports: [500] },
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/4Fy9rURbfF2bsl4BvYunUO/RFC-261-Search-Contexts?node-id=581%3A4754',
        },
    })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

const defaultProps: SearchContextMenuProps = {
    availableSearchContexts: [
        {
            __typename: 'SearchContext',
            id: '1',
            spec: 'global',
            autoDefined: true,
            description: 'All repositories on Sourcegraph',
        },
        {
            __typename: 'SearchContext',
            id: '2',
            spec: '@username',
            autoDefined: true,
            description: 'Your repositories on Sourcegraph',
        },
        {
            __typename: 'SearchContext',
            id: '3',
            spec: '@username/test-version-1.5',
            autoDefined: true,
            description: 'Only code in version 1.5',
        },
    ],
    defaultSearchContextSpec: 'global',
    selectedSearchContextSpec: 'global',
    selectSearchContextSpec: () => {},
    closeMenu: () => {},
}

add('default', () => <WebStory>{() => <SearchContextMenu {...defaultProps} />}</WebStory>, {})

add(
    'empty',
    () => <WebStory>{() => <SearchContextMenu {...defaultProps} availableSearchContexts={[]} />}</WebStory>,
    {}
)
