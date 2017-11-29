import MoonIcon from '@sourcegraph/icons/lib/Moon'
import SunIcon from '@sourcegraph/icons/lib/Sun'
import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'

interface Props {
    onToggleTheme: (isLightTheme: boolean) => void
    isLightTheme: boolean
}

interface State {}

export class ThemeSwitcher extends React.Component<Props, State> {
    public state: State = {}

    public render(): JSX.Element | null {
        return (
            <div className="theme-switcher theme-switcher__nav-bar">
                <div
                    className={
                        'theme-switcher__button' +
                        (this.props.isLightTheme
                            ? ' theme-switcher__button--selected theme-switcher__button--left'
                            : '')
                    }
                >
                    <a
                        className="btn btn-icon theme-switcher__link"
                        onClick={this.enableLightTheme}
                        title="Toggle theme"
                    >
                        <SunIcon />
                    </a>
                </div>
                <div
                    className={
                        'theme-switcher__button' +
                        (!this.props.isLightTheme
                            ? ' theme-switcher__button--selected theme-switcher__button--right'
                            : '')
                    }
                >
                    <a
                        className="btn btn-icon theme-switcher__link"
                        onClick={this.enableDarkTheme}
                        title="Toggle theme"
                    >
                        <MoonIcon />
                    </a>
                </div>
            </div>
        )
    }

    private enableLightTheme = () => {
        this.props.onToggleTheme(true)
        eventLogger.log('LightThemeClicked')
    }

    private enableDarkTheme = () => {
        this.props.onToggleTheme(false)
        eventLogger.log('DarkThemeClicked')
    }
}
