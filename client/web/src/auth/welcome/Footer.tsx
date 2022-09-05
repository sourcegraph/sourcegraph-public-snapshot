import React from 'react'

import { Button, Link } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import { FinishWelcomeFlow } from '../PostSignUpPage'
import { useSteps } from '../Steps/context'

interface Props {
    onFinish: FinishWelcomeFlow
    isSkippable?: boolean
}

export const Footer: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ onFinish, isSkippable }) => {
    const { setStep, currentIndex, currentStep } = useSteps()

    return (
        <div className="d-flex align-items-center justify-content-end mt-4 w-100">
            {!currentStep.isLastStep && (
                <Link
                    to="https://docs.sourcegraph.com/code_search/explanations/code_visibility_on_sourcegraph_cloud"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="mr-auto"
                >
                    <small>Who can see my code on Sourcegraph? </small>
                </Link>
            )}

            <div>
                {currentStep.isFirstStep && (
                    <Button
                        className="font-weight-normal"
                        onClick={event => {
                            event.currentTarget.blur()
                            setStep(currentIndex + 1)
                        }}
                        variant="link"
                    >
                        Not right now
                    </Button>
                )}

                <LoaderButton
                    alwaysShowLabel={true}
                    label={
                        isSkippable === true && !currentStep.isComplete
                            ? 'Skip this step'
                            : currentStep.isLastStep
                            ? 'Start searching'
                            : 'Continue'
                    }
                    className="float-right ml-2"
                    disabled={!currentStep.isComplete && !isSkippable}
                    variant="primary"
                    onClick={event => {
                        if (currentStep.isLastStep) {
                            onFinish(event, { eventName: 'StartSearching_Clicked' })
                        } else {
                            event.currentTarget.blur()
                            setStep(currentIndex + 1)
                        }
                    }}
                />
            </div>
        </div>
    )
}
