import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { LoaderButton } from '@sourcegraph/web/src/components/LoaderButton'

import { useSteps } from '../Steps/context'

interface Props {
    onFinish: () => void
}

export const Footer: React.FunctionComponent<Props> = ({ onFinish }) => {
    const { setStep, currentIndex, currentStep } = useSteps()

    return (
        <div className="mt-4">
            {!currentStep.isLastStep && (
                <Link
                    to="https://docs.sourcegraph.com/code_search/explanations/code_visibility_on_sourcegraph_cloud"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Who can see my code on Sourcegraph?{' '}
                </Link>
            )}

            <LoaderButton
                type="button"
                alwaysShowLabel={true}
                label={currentStep.isLastStep ? 'Start searching' : 'Continue'}
                className="btn btn-primary float-right ml-2"
                disabled={!currentStep.isComplete}
                onClick={currentStep.isLastStep ? onFinish : () => setStep(currentIndex + 1)}
            />

            {!currentStep.isLastStep && (
                <button
                    type="button"
                    className="btn btn-link font-weight-normal text-secondary float-right"
                    onClick={onFinish}
                >
                    Not right now
                </button>
            )}
        </div>
    )
}
