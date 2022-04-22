import { storiesOf } from '@storybook/react'
import { NEVER } from 'rxjs'
import sinon from 'sinon'

import { ISearchContext } from '@sourcegraph/shared/src/schema'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../components/WebStory'

import { DeleteSearchContextModal } from './DeleteSearchContextModal'

const { add } = storiesOf('web/enterprise/searchContexts/DeleteSearchContextModal', module)
    .addParameters({
        chromatic: { viewports: [1200], disableSnapshot: false },
    })
    .addDecorator(story => <div className="p-3 container">{story()}</div>)

const searchContext = {
    __typename: 'SearchContext',
    id: '1',
} as ISearchContext

add(
    'DeleteSearchContextModal',
    () => (
        <WebStory>
            {webProps => (
                <DeleteSearchContextModal
                    {...webProps}
                    isOpen={true}
                    searchContext={searchContext}
                    toggleDeleteModal={sinon.fake()}
                    deleteSearchContext={sinon.fake(() => NEVER)}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                />
            )}
        </WebStory>
    ),
    {}
)
