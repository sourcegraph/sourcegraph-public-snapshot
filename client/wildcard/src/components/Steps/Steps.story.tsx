import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Container } from '../Container'

import { Steps, Step } from './Steps'

const { add } = storiesOf('wildcard/Steps', module).addDecorator(story => (
    <BrandedStory styles={webStyles}>{() => <Container>{story()}</Container>}</BrandedStory>
))

add('Generic', () => {
    const [step, setStep] = useState(1)
    return (
        <Steps current={step} onChange={setStep} initial={1}>
            <Step title="Connect with code hosts" />
            <Step title="Add Repositories" />
            <Step title="Start Searching" />
        </Steps>
    )
})
