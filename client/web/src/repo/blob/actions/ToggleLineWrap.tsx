import * as React from 'react'

import WrapIcon from 'mdi-react/WrapIcon'
import { fromEvent, Subject, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'

import { WrapDisabledIcon } from '@sourcegraph/shared/src/components/icons'
import { TooltipController, Icon } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../tracking/eventLogger'
import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'
import { RepoHeaderContext } from '../../RepoHeader'

/**
 * A repository header action that toggles the line wrapping behavior for long lines in code files.
 */
export class ToggleLineWrap extends React.PureComponent<
    {
        /**
         * Called when the line wrapping behavior is toggled, with the new value (true means on,
         * false means off).
         */
        onDidUpdate: (value: boolean) => void
    } & RepoHeaderContext,
    { value: boolean }
> {
    private static STORAGE_KEY = 'wrap-code'

    public state = { value: ToggleLineWrap.getValue() }

    private updates = new Subject<boolean>()
    private subscriptions = new Subscription()

    /**
     * Reports the current line wrap behavior (true means on, false means off).
     */
    public static getValue(): boolean {
        return localStorage.getItem(ToggleLineWrap.STORAGE_KEY) === 'true' // default to off
    }

    /**
     * Sets the line wrap behavior (true means on, false means off).
     */
    private static setValue(value: boolean): void {
        localStorage.setItem(ToggleLineWrap.STORAGE_KEY, String(value))
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.updates.subscribe(value => {
                eventLogger.log(value ? 'WrappedCode' : 'UnwrappedCode')
                ToggleLineWrap.setValue(value)
                this.setState({ value })
                this.props.onDidUpdate(value)
                TooltipController.forceUpdate()
            })
        )

        // Toggle when the user presses 'alt+z'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                // Opt/alt+z shortcut
                .pipe(filter(event => event.altKey && event.code === 'KeyZ'))
                .subscribe(event => {
                    event.preventDefault()
                    this.updates.next(!this.state.value)
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.props.actionType === 'dropdown') {
            return (
                <RepoHeaderActionButtonLink file={true} onSelect={this.onClick}>
                    <Icon as={this.state.value ? WrapDisabledIcon : WrapIcon} />
                    <span>{this.state.value ? 'Disable' : 'Enable'} wrapping long lines (Alt+Z/Opt+Z)</span>
                </RepoHeaderActionButtonLink>
            )
        }

        return (
            <RepoHeaderActionButtonLink
                className="btn-icon"
                file={false}
                onSelect={this.onClick}
                data-tooltip={`${this.state.value ? 'Disable' : 'Enable'} wrapping long lines (Alt+Z/Opt+Z)`}
            >
                <Icon as={this.state.value ? WrapDisabledIcon : WrapIcon} />
            </RepoHeaderActionButtonLink>
        )
    }

    private onClick = (): void => this.updates.next(!this.state.value)
}
