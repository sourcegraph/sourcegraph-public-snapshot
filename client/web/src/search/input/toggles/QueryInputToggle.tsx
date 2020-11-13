import React from 'react'
import * as H from 'history'
import { Subscription, fromEvent } from 'rxjs'
import { filter } from 'rxjs/operators'
import { PatternTypeProps, CaseSensitivityProps } from '../..'
import { FiltersToTypeAndValue } from '../../../../../shared/src/search/interactive/util'
import classNames from 'classnames'
import { Key } from 'ts-key-enum'

export interface ToggleProps extends PatternTypeProps, CaseSensitivityProps {
    history: H.History
    /** Search query in the main query input. */
    navbarSearchQuery: string
    /** Title of the toggle.  */
    title: string
    /** Icon to display.  */
    icon: React.ComponentType<{ className?: string }>
    /** Condition for when the toggle should have an active state.  */
    isActive: boolean
    /** Callback on toggle.  */
    onToggle: () => void
    /**
     * A list of conditions to disable the toggle, displaying an associated tooltip when the condition is true.
     * For multiple true conditions, use the first rule that evalutes to true.
     */
    disabledRules?: { condition: boolean; reason: string }[]
    /** Filters in the query in interactive mode. */
    filtersInQuery?: FiltersToTypeAndValue
    hasGlobalQueryBehavior?: boolean
    className?: string
    activeClassName?: string
}

/**
 * A toggle displayed in the QueryInput.
 */
export class QueryInputToggle extends React.Component<ToggleProps> {
    private subscriptions = new Subscription()
    private toggleCheckbox = React.createRef<HTMLDivElement>()

    public componentDidMount(): void {
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(
                    filter(
                        event =>
                            document.activeElement === this.toggleCheckbox.current &&
                            (event.key === Key.Enter || event.key === ' ')
                    )
                )
                .subscribe(event => {
                    event.preventDefault()
                    this.props.onToggle()
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const Icon = this.props.icon

        let disabled = false
        let tooltipValue = `${this.props.isActive ? 'Disable' : 'Enable'} ${this.props.title.toLowerCase()}`
        if (this.props.disabledRules) {
            for (const rule of this.props.disabledRules) {
                if (rule.condition) {
                    disabled = true
                    tooltipValue = rule.reason
                    break
                }
            }
        }

        return (
            <div
                ref={this.toggleCheckbox}
                onClick={this.props.onToggle}
                className={classNames(
                    'btn btn-icon icon-inline toggle-container__toggle test-regexp-toggle',
                    this.props.className,
                    { disabled },
                    { 'toggle-container__toggle--active': this.props.isActive },
                    this.props.activeClassName
                )}
                role="checkbox"
                aria-disabled={disabled}
                aria-checked={this.props.isActive}
                aria-label={`${this.props.title} toggle`}
                tabIndex={0}
                data-tooltip={tooltipValue}
            >
                <Icon />
            </div>
        )
    }
}
