import React from 'react'

import { mdiAlertCircle, mdiCheckBold, mdiContentSave, mdiLinkVariantRemove, mdiTimerSand } from '@mdi/js'

import { Icon, IconType, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceStepFields } from '../../../../../graphql-operations'

interface StepStateIconProps {
    step: BatchSpecWorkspaceStepFields
}

const getProps = (step: BatchSpecWorkspaceStepFields): [classNameVariant: string, icon: IconType, label: string] => {
    if (step.cachedResultFound) {
        return ['text-success', mdiContentSave, 'A cached result for this step has been found.']
    }
    if (step.skipped) {
        return ['text-muted', mdiLinkVariantRemove, 'This step has been skipped.']
    }
    if (!step.startedAt) {
        return ['text-muted', mdiTimerSand, 'This step has not started yet.']
    }
    if (!step.finishedAt) {
        return ['text-muted', LoadingSpinner, 'This step is currently running.']
    }
    if (step.exitCode === 0) {
        return ['text-success', mdiCheckBold, 'This step finished running successfully.']
    }
    return ['text-danger', mdiAlertCircle, `This step failed with exit code ${String(step.exitCode)}.`]
}

export const StepStateIcon: React.FunctionComponent<React.PropsWithChildren<StepStateIconProps>> = ({ step }) => {
    const [classNameVariant, IconElement, label] = getProps(step)

    return (
        <div className="d-flex flex-shrink-0">
            <Tooltip content={label} placement="bottom">
                <Icon
                    className={classNameVariant}
                    aria-label={label}
                    {...(typeof IconElement === 'string' ? { svgPath: IconElement } : { as: IconElement })}
                />
            </Tooltip>
        </div>
    )
}
