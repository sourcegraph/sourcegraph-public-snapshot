import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { Container, Button } from '@sourcegraph/wildcard'

import { Steps, Step, StepList, StepPanels, StepPanel, StepActions, useSteps } from '.'

const Actions = () => {
    const { setStep, currentIndex, currentStep } = useSteps()
    return (
        <>
            <Button disabled={currentStep.isFirstStep} onClick={() => setStep(currentIndex - 1)} variant="primary">
                Previous
            </Button>
            <Button disabled={currentStep.isLastStep} onClick={() => setStep(currentIndex + 1)} variant="primary">
                Next
            </Button>
        </>
    )
}

export const Stepper: Story = () => (
    <BrandedStory styles={webStyles}>
        {() => (
            <Container>
                <Steps initialStep={2}>
                    <StepList numeric={true}>
                        <Step borderColor="blue">Panel 1 title</Step>
                        <Step borderColor="orange">Panel 2 Title</Step>
                        <Step borderColor="purple">Panel 3 Title</Step>
                    </StepList>
                    <StepPanels>
                        <StepPanel>Panel 1</StepPanel>
                        <StepPanel>Panel 2</StepPanel>
                        <StepPanel>Panel 3</StepPanel>
                    </StepPanels>
                    <StepActions>
                        <Actions />
                    </StepActions>
                </Steps>
            </Container>
        )}
    </BrandedStory>
)

Stepper.storyName = 'Steps component top navigation'

export default {
    title: 'web/Steps',
    component: Stepper,
} as Meta
