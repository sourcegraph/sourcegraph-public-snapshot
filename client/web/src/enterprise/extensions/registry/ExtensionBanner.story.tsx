import { storiesOf } from '@storybook/react'
import React from 'react'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { ExtensionBanner } from './ExtensionBanner'

const { add } = storiesOf('web/Extensions', module).addDecorator(story => <div className="p-4">{story()}</div>)

add('ExtensionBanner', () => <EnterpriseWebStory>{() => <ExtensionBanner />}</EnterpriseWebStory>, {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=420%3A10',
    },
})
