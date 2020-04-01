import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Accordion, AccordionItem, AccordionButton, AccordionPanel } from '@reach/accordion'
import H from 'history'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import * as React from 'react'
import { Link } from '../Link'
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
export class ActivationChecklistItem extends React.PureComponent<ActivationChecklistItemProps, {}> {
    private onClick = (e: React.MouseEvent<HTMLElement>): void => {
        if (this.props.onClick) {
            this.props.onClick(e, this.props.history)
        }
    }

    public render(): JSX.Element {
        const checkboxElem = (
            <div>
                <span className="icon-down">
                    <ChevronDownIcon />
                </span>
                <span className="icon-right">
                    <ChevronRightIcon />
                </span>
                {this.props.done ? (
                    <CheckCircleIcon className="icon-inline text-success" />
                ) : (
                    <CheckboxBlankCircleIcon className="icon-inline text-muted" />
                )}
                <span className="mr-2 ml-2">{this.props.title}</span>
            </div>
        )

        return (
            <>{checkboxElem}</>

            // <div onClick={this.onClick} data-tooltip={this.props.detail}>
            //     <button type="button" className="btn btn-link text-left w-100 p-0 border-0">
            //         {checkboxElem}
            //     </button>
            // </div>
        )
    }
}

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
            <div className={`list-group list-group-flush ${this.props.className || ''}`}>
                <Accordion collapsible={true}>
                    {this.props.steps.map(step => (
                        <AccordionItem key={step.id}>
                            <AccordionButton className="list-group-item list-group-item-action">
                                <ActivationChecklistItem
                                    key={step.id}
                                    {...step}
                                    history={this.props.history}
                                    done={(this.props.completed && this.props.completed[step.id]) || false}
                                />
                            </AccordionButton>
                            <AccordionPanel>{step.detail}</AccordionPanel>
                        </AccordionItem>
                    ))}
                </Accordion>
            </div>
        ) : (
            <LoadingSpinner className="icon-inline my-2" />
        )
    }
}
