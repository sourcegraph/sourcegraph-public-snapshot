import { Accordion, AccordionItem, AccordionButton, AccordionPanel } from '@reach/accordion'
import classNames from 'classnames'
import CheckboxBlankCircleOutlineIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import * as React from 'react'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import { ActivationCompletionStatus, ActivationStep } from './Activation'
import styles from './ActivationChecklist.module.scss'

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
    <div className={classNames('d-flex justify-content-between', styles.activationChecklistItem, className)}>
        <div className="d-flex align-items-center">
            <span className={classNames('icon-inline', styles.iconContainer, styles.iconDown)}>
                <ChevronDownIcon className={styles.icon} />
            </span>
            <span className={classNames('icon-inline', styles.iconContainer, styles.iconRight)}>
                <ChevronRightIcon className={styles.icon} />
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
        return <LoadingSpinner className="my-2" />
    }

    return (
        <div className={classNames('list-group list-group-flush', styles.activationChecklist, className)}>
            <Accordion collapsible={true}>
                {steps.map(step => (
                    <AccordionItem key={step.id} className={classNames('list-group-item', styles.container)}>
                        <AccordionButton
                            className={classNames('list-group-item list-group-item-action btn-link', styles.button)}
                        >
                            <ActivationChecklistItem
                                key={step.id}
                                {...step}
                                done={completed?.[step.id] || false}
                                className={buttonClassName}
                            />
                        </AccordionButton>
                        <AccordionPanel className="px-2">
                            <div className={classNames('pb-1', styles.detail)}>{step.detail}</div>
                        </AccordionPanel>
                    </AccordionItem>
                ))}
            </Accordion>
        </div>
    )
}
