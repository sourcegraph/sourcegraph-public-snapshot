import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Accordion, AccordionItem, AccordionButton, AccordionPanel } from '@reach/accordion'
import H from 'history'
import CheckboxBlankCircleOutlineIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import * as React from 'react'
import { ActivationCompletionStatus, ActivationStep } from './Activation'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

interface ActivationChecklistItemProps extends ActivationStep {
    done: boolean
    history: H.History
}

/**
 * A single item in the activation checklist.
 */
export const ActivationChecklistItem: React.FunctionComponent<ActivationChecklistItemProps> = (
    props: ActivationChecklistItemProps
) => (
    <div className="d-flex justify-content-between py-1 px-2">
        <div>
            <span className="icon-inline icon-down">
                <ChevronDownIcon />
            </span>
            <span className="icon-inline icon-right">
                <ChevronRightIcon />
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
                                    history={this.props.history}
                                    done={(this.props.completed && this.props.completed[step.id]) || false}
                                />
                                <AccordionPanel className="px-2">
                                    <div className="activation-checklist__detail pl-2 pb-1">{step.detail}</div>
                                </AccordionPanel>
                            </AccordionButton>
                        </AccordionItem>
                    ))}
                </Accordion>
            </div>
        ) : (
            <LoadingSpinner className="icon-inline my-2" />
        )
    }
}
