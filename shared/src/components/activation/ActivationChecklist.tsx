import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import * as React from 'react'
import { Link } from '../Link'
import { ActivationCompletionStatus, ActivationStep } from './Activation'

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
            <div className="d-flex justify-content-between">
                <span className="mr-2">{this.props.title}</span>
                {this.props.done ? (
                    <CheckIcon className="icon-inline text-success" />
                ) : (
                    <CheckboxBlankCircleIcon className="icon-inline text-muted" />
                )}
            </div>
        )

        return (
            <div onClick={this.onClick} data-tooltip={this.props.detail}>
                {this.props.link ? (
                    <Link {...this.props.link}>{checkboxElem}</Link>
                ) : (
                    <button type="button" className="btn btn-link text-left w-100 p-0">
                        {checkboxElem}
                    </button>
                )}
            </div>
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
                {this.props.steps.map(step => (
                    <div key={step.id} className="list-group-item list-group-item-action">
                        <ActivationChecklistItem
                            {...step}
                            history={this.props.history}
                            done={(this.props.completed && this.props.completed[step.id]) || false}
                        />
                    </div>
                ))}
            </div>
        ) : (
            <LoadingSpinner className="icon-inline my-2" />
        )
    }
}
