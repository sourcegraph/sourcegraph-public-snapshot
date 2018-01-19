import MoonIcon from '@sourcegraph/icons/lib/Moon'
import SunIcon from '@sourcegraph/icons/lib/Sun'
import * as React from 'react'
import { Tooltip } from './tooltip/Tooltip'

interface Props {
    isLightTheme: boolean
    onThemeChange: () => void
}

export class ThemeSwitcher extends React.PureComponent<Props, {}> {
    public componentDidMount(): void {
        Tooltip.forceUpdate()
    }

    public componentDidUpdate(): void {
        Tooltip.forceUpdate()
    }

    public render(): JSX.Element | null {
        return (
            <div
                className="theme-switcher theme-switcher__nav-bar"
                onClick={this.props.onThemeChange}
                data-tooltip={this.props.isLightTheme ? 'Switch to dark color theme' : 'Switch to light color theme'}
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
