import React from 'react'

import { LoaderButton } from '@sourcegraph/web/src/components/LoaderButton'
import { RouterLink } from '@sourcegraph/wildcard'

import { FinishWelcomeFlow } from '../PostSignUpPage'
import { useSteps } from '../Steps/context'

interface Props {
    onFinish: FinishWelcomeFlow
}

export const Footer: React.FunctionComponent<Props> = ({ onFinish }) => {
    const { setStep, currentIndex, currentStep } = useSteps()

    return (
        <div className="d-flex align-items-center justify-content-end mt-4">
            {!currentStep.isLastStep && (
                <RouterLink
                    to="https://docs.sourcegraph.com/code_search/explanations/code_visibility_on_sourcegraph_cloud"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="mr-auto"
                >
                    <small>Who can see my code on Sourcegraph? </small>
                </RouterLink>
            )}

            <div>
                {!currentStep.isLastStep && (
                    <button
                        type="button"
                        className="btn btn-link font-weight-normal text-secondary"
                        onClick={event =>
                            onFinish(event, { eventName: 'NotRightNow_Clicked', tabNumber: currentIndex })
                        }
                    >
                        Not right now
                    </button>
                )}
                <LoaderButton
                    type="button"
                    alwaysShowLabel={true}
                    label={currentStep.isLastStep ? 'Start searching' : 'Continue'}
                    className="btn btn-primary float-right ml-2"
                    disabled={!currentStep.isComplete}
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
