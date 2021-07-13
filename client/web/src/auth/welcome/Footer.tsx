import React from 'react'

import { useSteps } from '@sourcegraph/wildcard/src/components/Steps'

export const Footer: React.FunctionComponent = () => {
    // const steps = useSteps()

    console.log('steps')

    return <div>footer</div>

    // return (<div className="mt-4">
    // <button
    //     type="button"
    //     className="btn btn-primary float-right ml-2"
    //     disabled={!!externalServices && externalServices?.length === 0}
    //     // disabled={!isCurrentStepComplete()}
    //     onClick={isLastStep ? goToSearch : goToNextTab}
    // >
    //     {isLastStep ? 'Start searching' : 'Continue'}
    // </button>

    // {!isLastStep && (
    //     <button
    //         type="button"
    //         className="btn btn-link font-weight-normal text-secondary float-right"
    //         onClick={skipPostSignup}
    //     >
    //         Not right now
    //     </button>
    // )}
}
