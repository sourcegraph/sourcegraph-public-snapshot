import MoonIcon from '@sourcegraph/icons/lib/Moon'
import SunIcon from '@sourcegraph/icons/lib/Sun'
import * as React from 'react'

interface Props {
    isLightTheme: boolean
    className?: string
    onThemeChange: () => void
}

export const ThemeSwitcher: React.SFC<Props> = props => (
    <div
        className={`theme-switcher theme-switcher__nav-bar ${props.className || ''}`}
        onClick={props.onThemeChange}
        title={props.isLightTheme ? 'Switch to dark color theme' : 'Switch to light color theme'}
    >
        <button className="theme-switcher__button btn-icon">{props.isLightTheme ? <MoonIcon /> : <SunIcon />}</button>
    </div>
)
