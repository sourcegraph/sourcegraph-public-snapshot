import { storiesOf } from '@storybook/react'
import { createMemoryHistory } from 'history'
import React from 'react'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { WebStory } from '../../../../components/WebStory'
import { authUser } from '../../../../search/panels/utils'

import { CreationSearchInsightPage, CreationSearchInsightPageProps } from './CreationSearchInsightPage'

const { add } = storiesOf('web/insights/CreationSearchInsightPage', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const PLATFORM_CONTEXT: CreationSearchInsightPageProps['platformContext'] = {
    // eslint-disable-next-line @typescript-eslint/require-await
    updateSettings: async (...args) => {
        console.log('PLATFORM CONTEXT update settings with', { ...args })
    },
}

const history = createMemoryHistory()

add('Page', () => (
    <CreationSearchInsightPage
        history={history}
        platformContext={PLATFORM_CONTEXT}
        settingsCascade={EMPTY_SETTINGS_CASCADE}
        authenticatedUser={authUser}
    />
))
