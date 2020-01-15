import WrapIcon from 'mdi-react/WrapIcon'
import * as React from 'react'
import { fromEvent, Subject, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import { WrapDisabledIcon } from '../../../../../shared/src/components/icons'
import { LinkOrButton } from '../../../../../shared/src/components/LinkOrButton'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { eventLogger } from '../../../tracking/eventLogger'

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
    },
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
        that.subscriptions.add(
            that.updates.subscribe(value => {
                eventLogger.log(value ? 'WrappedCode' : 'UnwrappedCode')
                ToggleLineWrap.setValue(value)
                that.setState({ value })
                that.props.onDidUpdate(value)
                Tooltip.forceUpdate()
            })
        )

        // Toggle when the user presses 'alt+z'.
        that.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                // Opt/alt+z shortcut
                .pipe(filter(event => event.altKey && event.key === 'z'))
                .subscribe(event => {
                    event.preventDefault()
                    that.updates.next(!that.state.value)
                })
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <LinkOrButton
                onSelect={that.onClick}
                data-tooltip={`${that.state.value ? 'Disable' : 'Enable'} wrapping long lines (Alt+Z/Opt+Z)`}
            >
                {that.state.value ? <WrapDisabledIcon className="icon-inline" /> : <WrapIcon className="icon-inline" />}
            </LinkOrButton>
        )
    }

    private onClick = (): void => that.updates.next(!that.state.value)
}
