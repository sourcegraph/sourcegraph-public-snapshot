import React from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { Icon, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceStepFields } from '../../../../../graphql-operations'

interface StepStateIconProps {
    step: BatchSpecWorkspaceStepFields
}

const getProps = (
    step: BatchSpecWorkspaceStepFields
): [classNameVariant: string, icon: React.ElementType, label: string] => {
    if (step.cachedResultFound) {
        return ['text-success', ContentSaveIcon, 'A cached result for this step has been found.']
    }
    if (step.skipped) {
        return ['text-muted', LinkVariantRemoveIcon, 'This step has been skipped.']
    }
    if (!step.startedAt) {
        return ['text-muted', TimerSandIcon, 'This step has not started yet.']
    }
    if (!step.finishedAt) {
        return ['text-muted', LoadingSpinner, 'This step is currently running.']
    }
    if (step.exitCode === 0) {
        return ['text-success', CheckBoldIcon, 'This step finished running successfully.']
    }
    return ['text-danger', AlertCircleIcon, `This step failed with exit code ${String(step.exitCode)}.`]
}

export const StepStateIcon: React.FunctionComponent<React.PropsWithChildren<StepStateIconProps>> = ({ step }) => {
    const [classNameVariant, IconElement, label] = getProps(step)

    return (
        <div className="d-flex flex-shrink-0">
            <Tooltip content={label} placement="bottom">
                <span>
                    <Icon className={classNameVariant} aria-label={label} as={IconElement} />
                </span>
            </Tooltip>
        </div>
    )
}
