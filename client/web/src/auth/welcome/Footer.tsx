import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { LoaderButton } from '@sourcegraph/web/src/components/LoaderButton'
import { Button } from '@sourcegraph/wildcard'

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
                {!currentStep.isLastStep && (
                    <Button
                        className="font-weight-normal text-secondary"
                        onClick={event =>
                            onFinish(event, { eventName: 'NotRightNow_Clicked', tabNumber: currentIndex })
                        }
                        variant="link"
                    >
                        Not right now
                    </Button>
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
