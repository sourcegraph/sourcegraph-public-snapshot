import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { WebStory } from '../components/WebStory'

import { ExperimentalSignUpPage } from './ExperimentalSignUpPage'

const { add } = storiesOf('web/auth/ExperimentalSignUpPage', module)

add('default', () => (
    <WebStory>
        {({ isLightTheme }) => (
            <ExperimentalSignUpPage isLightTheme={isLightTheme} source="test" onSignUp={sinon.stub()} />
        )}
    </WebStory>
))
