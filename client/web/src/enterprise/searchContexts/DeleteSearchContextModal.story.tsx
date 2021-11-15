import { storiesOf } from '@storybook/react'
import React from 'react'
import { NEVER } from 'rxjs'
import sinon from 'sinon'

import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'

import { WebStory } from '../../components/WebStory'

import { DeleteSearchContextModal } from './DeleteSearchContextModal'

const { add } = storiesOf('web/searchContexts/DeleteSearchContextModal', module)
    .addParameters({
        chromatic: { viewports: [1200] },
    })
    .addDecorator(story => <div className="p-3 container">{story()}</div>)

const searchContext = {
    __typename: 'SearchContext',
    id: '1',
} as ISearchContext

add(
    'delete modal',
    () => (
        <WebStory>
            {webProps => (
                <DeleteSearchContextModal
                    {...webProps}
                    isOpen={true}
                    searchContext={searchContext}
                    toggleDeleteModal={sinon.fake()}
                    deleteSearchContext={sinon.fake(() => NEVER)}
                />
            )}
        </WebStory>
    ),
    {}
)
