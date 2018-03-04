import MoonIcon from '@sourcegraph/icons/lib/Moon'
import SunIcon from '@sourcegraph/icons/lib/Sun'
import * as React from 'react'
import { Tooltip } from './tooltip/Tooltip'

interface Props {
    isLightTheme: boolean
    className?: string
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
                className={`theme-switcher theme-switcher__nav-bar ${this.props.className || ''}`}
                onClick={this.props.onThemeChange}
                data-tooltip={this.props.isLightTheme ? 'Switch to dark color theme' : 'Switch to light color theme'}
            >
                <div className="theme-switcher__button btn-icon">
                    {this.props.isLightTheme ? <MoonIcon /> : <SunIcon />}
                </div>
            </div>
        )
    }
}
