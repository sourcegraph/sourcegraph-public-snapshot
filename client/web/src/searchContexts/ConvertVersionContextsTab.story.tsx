import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../components/WebStory'

import { ConvertVersionContextsTabProps, ConvertVersionContextsTab } from './ConvertVersionContextsTab'

const { add } = storiesOf('web/searchContexts/ConvertVersionContextsTab', module)
    .addParameters({
        chromatic: { viewports: [500] },
    })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

const defaultProps: ConvertVersionContextsTabProps = {
    availableVersionContexts: [
        {
            name: 'version context 1',
            description: 'description 1',
            revisions: [],
        },
        {
            name: 'version context 2',
            description: 'description 2',
            revisions: [],
        },
        {
            name: 'version context 3',
            description: 'description 3',
            revisions: [],
        },
    ],
    isSearchContextSpecAvailable: () => of(false),
    convertVersionContextToSearchContext: (name: string) => of({ id: name, spec: name }),
}

add('default', () => <WebStory>{() => <ConvertVersionContextsTab {...defaultProps} />}</WebStory>, {})

add(
    'all converted',
    () => (
        <WebStory>
            {() => <ConvertVersionContextsTab {...defaultProps} isSearchContextSpecAvailable={() => of(true)} />}
        </WebStory>
    ),
    {}
)
