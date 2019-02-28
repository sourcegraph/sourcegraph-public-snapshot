import H from 'history'
import CheckboxBlankCircleOutlineIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CheckboxMarkedCircleOutlineIcon from 'mdi-react/CheckboxMarkedCircleOutlineIcon'
import * as React from 'react'
import { ActivationStep } from './Activation'

interface ActivationChecklistItemProps extends ActivationStep {
    done: boolean
    history: H.History
}

/**
 * A single item in the activation checklist.
 */
export class ActivationChecklistItem extends React.PureComponent<ActivationChecklistItemProps, {}> {
    private doAction = () => this.props.action(this.props.history)

    public render(): JSX.Element {
        return (
            <div className="activation-item" onClick={this.doAction} data-tooltip={this.props.detail}>
                {this.props.done ? (
                    <CheckboxMarkedCircleOutlineIcon className="icon-inline activation-item__checkbox--done" />
                ) : (
                    <CheckboxBlankCircleOutlineIcon className="icon-inline activation-item__checkbox--todo" />
                )}
                &nbsp;&nbsp;
                {this.props.title}
                &nbsp;
            </div>
        )
    }
}

export interface ActivationChecklistProps {
    history: H.History
    steps: ActivationStep[]
    completed?: { [key: string]: boolean }
}

/**
 * Renders an activation checklist.
 */
export class ActivationChecklist extends React.PureComponent<ActivationChecklistProps, {}> {
    public render(): JSX.Element {
        return (
            <div className="activation-checklist">
                {this.props.completed ? (
                    this.props.steps.map(s => (
                        <div key={s.id} className="activation-checklist__item">
                            <ActivationChecklistItem
                                {...s}
                                history={this.props.history}
                                done={(this.props.completed && this.props.completed[s.id]) || false}
                            />
                        </div>
                    ))
                ) : (
                    <div>Loading...</div>
                )}
            </div>
        )
    }
}
