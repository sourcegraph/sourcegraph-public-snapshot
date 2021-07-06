import { Accordion, AccordionItem, AccordionButton, AccordionPanel } from '@reach/accordion'
import classNames from 'classnames'
import CheckboxBlankCircleOutlineIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import * as React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { ActivationCompletionStatus, ActivationStep } from './Activation'

interface ActivationChecklistItemProps extends ActivationStep {
    done: boolean
    className?: string
}

/**
 * A single item in the activation checklist.
 */
export const ActivationChecklistItem: React.FunctionComponent<ActivationChecklistItemProps> = ({
    className = '',
    ...props
}: ActivationChecklistItemProps) => (
    <div className={classNames('activation-checklist-item d-flex justify-content-between', className)}>
        <div className="d-flex align-items-center">
            <span className="activation-checklist-item__icon-container icon-inline icon-down">
                <ChevronDownIcon className="activation-checklist-item__icon" />
            </span>
            <span className="activation-checklist-item__icon-container icon-inline icon-right">
                <ChevronRightIcon className="activation-checklist-item__icon" />
            </span>
            <span>{props.title}</span>
        </div>
        <div>
            {props.done ? (
                <CheckCircleIcon className="icon-inline text-success" />
            ) : (
                <CheckboxBlankCircleOutlineIcon className="icon-inline text-muted" />
            )}
        </div>
    </div>
)

export interface ActivationChecklistProps {
    steps: ActivationStep[]
    completed?: ActivationCompletionStatus
    className?: string
    buttonClassName?: string
}

/**
 * Renders an activation checklist.
 */
export const ActivationChecklist: React.FunctionComponent<ActivationChecklistProps> = ({
    className,
    steps,
    completed,
    buttonClassName,
}) => {
    if (!completed) {
        return <LoadingSpinner className="icon-inline my-2" />
    }

    return (
        <div className={`activation-checklist list-group list-group-flush ${className || ''}`}>
            <Accordion collapsible={true}>
                {steps.map(step => (
                    <AccordionItem key={step.id} className="activation-checklist__container list-group-item">
                        <AccordionButton className="activation-checklist__button list-group-item list-group-item-action btn-link">
                            <ActivationChecklistItem
                                key={step.id}
                                {...step}
                                done={completed?.[step.id] || false}
                                className={buttonClassName}
                            />
                        </AccordionButton>
                        <AccordionPanel className="px-2">
                            <div className="activation-checklist__detail pb-1">{step.detail}</div>
                        </AccordionPanel>
                    </AccordionItem>
                ))}
            </Accordion>
        </div>
    )
}
