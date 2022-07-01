import * as React from 'react'

import { mdiChevronDown, mdiChevronRight, mdiCheckCircle, mdiCheckboxBlankCircleOutline } from '@mdi/js'
import { Accordion, AccordionItem, AccordionButton, AccordionPanel } from '@reach/accordion'
import classNames from 'classnames'

import { Button, LoadingSpinner, Icon } from '@sourcegraph/wildcard'

import { ActivationCompletionStatus, ActivationStep } from './Activation'

import styles from './ActivationChecklist.module.scss'

interface ActivationChecklistItemProps extends ActivationStep {
    done: boolean
    className?: string
}

/**
 * A single item in the activation checklist.
 */
export const ActivationChecklistItem: React.FunctionComponent<
    React.PropsWithChildren<ActivationChecklistItemProps>
> = ({ className = '', ...props }: ActivationChecklistItemProps) => (
    <div className={classNames('d-flex justify-content-between', styles.activationChecklistItem, className)}>
        <div className="d-flex align-items-center">
            <span className={styles.iconContainer}>
                <Icon
                    className={classNames(styles.icon, styles.iconDown)}
                    aria-hidden={true}
                    svgPath={mdiChevronDown}
                />
                <Icon
                    className={classNames(styles.icon, styles.iconRight)}
                    aria-hidden={true}
                    svgPath={mdiChevronRight}
                />
            </span>
            <span>{props.title}</span>
        </div>
        <div>
            {props.done ? (
                <Icon className="text-success" aria-label="Completed" svgPath={mdiCheckCircle} />
            ) : (
                <Icon className="text-muted" aria-label="Not completed" svgPath={mdiCheckboxBlankCircleOutline} />
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
export const ActivationChecklist: React.FunctionComponent<React.PropsWithChildren<ActivationChecklistProps>> = ({
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
                        <Button
                            as={AccordionButton}
                            variant="link"
                            className={classNames('list-group-item list-group-item-action', styles.button)}
                        >
                            <ActivationChecklistItem
                                key={step.id}
                                {...step}
                                done={completed?.[step.id] || false}
                                className={buttonClassName}
                            />
                        </Button>
                        <AccordionPanel className="px-2">
                            <div className={classNames('pb-1', styles.detail)}>{step.detail}</div>
                        </AccordionPanel>
                    </AccordionItem>
                ))}
            </Accordion>
        </div>
    )
}
