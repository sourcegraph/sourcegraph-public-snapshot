import React, { FunctionComponent } from 'react'
import { DockerStep, DockerStepProps } from './DockerStep'

export interface DockerStepsProps {
    steps: DockerStepProps['step'][]
    className?: string
}

export const DockerSteps: FunctionComponent<DockerStepsProps> = ({ steps, className }) => (
    <>
        <h3>Steps</h3>

        <div className={className}>
            <ul className="list-group">
                {steps.map(step => (
                    <DockerStep step={step} key={`${step.root}${step.image}${step.commands.join(' ')}`} />
                ))}
            </ul>
        </div>
    </>
)
