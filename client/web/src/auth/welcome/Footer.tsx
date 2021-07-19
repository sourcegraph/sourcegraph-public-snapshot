import React from 'react'
import { useHistory, useLocation } from 'react-router'

import { LoaderButton } from '@sourcegraph/web/src/components/LoaderButton'
import { useSteps } from '@sourcegraph/wildcard/src/components/Steps'

import { getReturnTo } from '../SignInSignUpCommon'

export const Footer: React.FunctionComponent = () => {
    const history = useHistory()
    const location = useLocation()
    const { setStep, currentIndex, steps, currentStep } = useSteps()

    const goToSearch = (): void => history.push(getReturnTo(location))

    console.log('steps', { setStep }, { currentIndex }, { steps }, { currentStep })

    return (
        <div className="mt-4">
            <LoaderButton
                type="button"
                alwaysShowLabel={true}
                label={currentStep.isLastStep ? 'Start searching' : 'Continue'}
                className="btn btn-primary float-right ml-2"
                disabled={!currentStep.isComplete}
                onClick={currentStep.isLastStep ? goToSearch : () => setStep(currentIndex + 1)}
            />

            {!currentStep.isLastStep && (
                <button
                    type="button"
                    className="btn btn-link font-weight-normal text-secondary float-right"
                    onClick={goToSearch}
                >
                    Not right now
                </button>
            )}
        </div>
    )
}
