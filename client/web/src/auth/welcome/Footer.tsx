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
                    <button type="button" className="btn btn-link font-weight-normal text-secondary" onClick={onFinish}>
                        Not right now
                    </button>
                )}

                <LoaderButton
                    type="button"
                    alwaysShowLabel={true}
                    label={currentStep.isLastStep ? 'Start searching' : 'Continue'}
                    className="btn btn-primary ml-2"
                    disabled={!currentStep.isComplete}
                    onClick={currentStep.isLastStep ? onFinish : () => setStep(currentIndex + 1)}
                />
            </div>
        </div>
    )
}
