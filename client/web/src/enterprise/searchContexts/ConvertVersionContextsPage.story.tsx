import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { EnterpriseWebStory } from '../components/EnterpriseWebStory'

import { ConvertVersionContextsPageProps, ConvertVersionContextsPage } from './ConvertVersionContextsPage'

const { add } = storiesOf('web/searchContexts/ConvertVersionContextsPage', module)
    .addParameters({
        chromatic: { viewports: [500] },
    })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

const defaultProps: ConvertVersionContextsPageProps = {
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

add(
    'default',
    () => <EnterpriseWebStory>{() => <ConvertVersionContextsPage {...defaultProps} />}</EnterpriseWebStory>,
    {}
)

add(
    'all converted',
    () => (
        <EnterpriseWebStory>
            {() => <ConvertVersionContextsPage {...defaultProps} isSearchContextSpecAvailable={() => of(true)} />}
        </EnterpriseWebStory>
    ),
    {}
)
