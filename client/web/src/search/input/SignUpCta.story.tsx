import { storiesOf } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'

import { SignUpCta } from './SignUpCta'

const { add } = storiesOf('web/SignUpCta', module).addDecorator(story => <div className="p-4">{story()}</div>)

add('SignUpCta', () => <WebStory>{() => <SignUpCta telemetryService={NOOP_TELEMETRY_SERVICE} />}</WebStory>, {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/F6xqBsBLJSUx3xY5zBOFg6/Homepage-concepts?node-id=100%3A10749',
    },
})
