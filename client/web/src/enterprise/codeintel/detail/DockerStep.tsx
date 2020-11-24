import React, { FunctionComponent } from 'react'

export interface DockerStepProps {
    step: {
        root: string
        image: string
        commands: string[]
    }
}

export const DockerStep: FunctionComponent<DockerStepProps> = ({ step }) => (
    <li className="list-group-item">
        <code>
            <strong>{step.image}</strong> {step.commands.join(' ')}
        </code>
        <span className="float-right">
            <span className="text-muted">executed in</span> /{step.root}
        </span>
    </li>
)
