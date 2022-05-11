import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { Container, Button } from '@sourcegraph/wildcard'

import { Steps, Step, StepList, StepPanels, StepPanel, useSteps } from '.'

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
                <Steps initialStep={2} totalSteps={3}>
                    <StepList numeric={true}>
                        <Step borderColor="blue">Panel 1 title</Step>
                        <Step borderColor="orange">Panel 2 Title</Step>
                        <Step borderColor="purple">Panel 3 Title</Step>
                    </StepList>
                    <StepPanels>
                        <StepPanel>
                            <div>Panel 1</div>
                            <Actions />
                        </StepPanel>
                        <StepPanel>
                            <div>Panel 2</div>
                            <Actions />
                        </StepPanel>
                        <StepPanel>
                            <div>Panel 3</div>
                            <Actions />
                        </StepPanel>
                    </StepPanels>
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
