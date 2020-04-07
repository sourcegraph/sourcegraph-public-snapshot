import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Accordion, AccordionItem, AccordionButton, AccordionPanel } from '@reach/accordion'
import H from 'history'
import CheckboxBlankCircleOutlineIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import * as React from 'react'
import { ActivationCompletionStatus, ActivationStep } from './Activation'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import classNames from 'classnames'

interface ActivationChecklistItemProps extends ActivationStep {
    done: boolean
    history: H.History
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
        <div>
            <span className="icon-inline icon-down">
                <ChevronDownIcon className="activation-checklist-item__icon" />
            </span>
            <span className="icon-inline icon-right">
                <ChevronRightIcon className="activation-checklist-item__icon" />
            </span>
            <span className="activation-checklist__title">{props.title}</span>
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
    history: H.History
    steps: ActivationStep[]
    completed?: ActivationCompletionStatus
    className?: string
}

/**
 * Renders an activation checklist.
 */
export class ActivationChecklist extends React.PureComponent<ActivationChecklistProps, {}> {
    public render(): JSX.Element {
        return this.props.completed ? (
            <div className={`activation-checklist list-group list-group-flush ${this.props.className || ''}`}>
                <Accordion collapsible={true}>
                    {this.props.steps.map(step => (
                        <AccordionItem key={step.id} className="list-group-item activation-checklist__item">
                            <AccordionButton className="list-group-item list-group-item-action activation-checklist__button">
                                <ActivationChecklistItem
                                    key={step.id}
                                    {...step}
                                    className="activation-checklist-item--admin-page"
                                    history={this.props.history}
                                    done={(this.props.completed && this.props.completed[step.id]) || false}
                                />
                            </AccordionButton>
                            <AccordionPanel className="px-2">
                                <div
                                    className="activation-checklist__detail pb-1"
                                    dangerouslySetInnerHTML={{ __html: step.detail }}
                                />
                            </AccordionPanel>
                        </AccordionItem>
                    ))}
                </Accordion>
            </div>
        ) : (
            <LoadingSpinner className="icon-inline my-2" />
        )
    }
}
