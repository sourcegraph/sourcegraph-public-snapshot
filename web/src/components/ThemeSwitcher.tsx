import MoonIcon from '@sourcegraph/icons/lib/Moon'
import SunIcon from '@sourcegraph/icons/lib/Sun'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { colorTheme, getColorTheme, setColorTheme } from '../settings/theme'

interface Props {}

interface State {
    isLightTheme: boolean
}

export class ThemeSwitcher extends React.Component<Props, State> {
    public state: State = { isLightTheme: getColorTheme() === 'light' }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(colorTheme.subscribe(theme => this.setState({ isLightTheme: theme === 'light' })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div
                className="theme-switcher theme-switcher__nav-bar"
                onClick={this.toggleTheme}
                title="Switch color theme"
            >
                <div
                    className={
                        'theme-switcher__button' +
                        (this.state.isLightTheme
                            ? ' theme-switcher__button--selected theme-switcher__button--left'
                            : '')
                    }
                >
                    <span className="btn-icon theme-switcher__link">
                        <SunIcon />
                    </span>
                </div>
                <div
                    className={
                        'theme-switcher__button' +
                        (!this.state.isLightTheme
                            ? ' theme-switcher__button--selected theme-switcher__button--right'
                            : '')
                    }
                >
                    <span className="btn-icon theme-switcher__link">
                        <MoonIcon />
                    </span>
                </div>
            </div>
        )
    }

    private toggleTheme = () => setColorTheme(getColorTheme() === 'light' ? 'dark' : 'light')
}
