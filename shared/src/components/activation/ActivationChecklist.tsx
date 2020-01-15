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
        if (that.props.onClick) {
            that.props.onClick(e, that.props.history)
        }
    }

    public render(): JSX.Element {
        const checkboxElem = (
            <div>
                {that.props.done ? (
                    <CheckIcon className="icon-inline text-success" />
                ) : (
                    <CheckboxBlankCircleIcon className="icon-inline text-muted" />
                )}
                <span className="mr-2 ml-2">{that.props.title}</span>
            </div>
        )

        return (
            <div onClick={that.onClick} data-tooltip={that.props.detail}>
                {that.props.link ? (
                    <Link {...that.props.link}>{checkboxElem}</Link>
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
        return that.props.completed ? (
            <div className={`list-group list-group-flush ${that.props.className || ''}`}>
                {that.props.steps.map(step => (
                    <div key={step.id} className="list-group-item list-group-item-action">
                        <ActivationChecklistItem
                            {...step}
                            history={that.props.history}
                            done={(that.props.completed && that.props.completed[step.id]) || false}
                        />
                    </div>
                ))}
            </div>
        ) : (
            <LoadingSpinner className="icon-inline my-2" />
        )
    }
}
