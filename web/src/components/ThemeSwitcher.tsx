import MoonIcon from '@sourcegraph/icons/lib/Moon'
import SunIcon from '@sourcegraph/icons/lib/Sun'
import * as React from 'react'

interface Props {
    onToggleTheme: () => void
    isLightTheme: boolean
}

interface State {}

export class ThemeSwitcher extends React.Component<Props, State> {
    public state: State = {}

    public render(): JSX.Element | null {
        return (
            <div
                className="theme-switcher theme-switcher__nav-bar"
                onClick={this.props.onToggleTheme}
                title="Switch color theme"
            >
                <div
                    className={
                        'theme-switcher__button' +
                        (this.props.isLightTheme
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
                        (!this.props.isLightTheme
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
}
