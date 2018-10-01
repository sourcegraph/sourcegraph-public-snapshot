import ThemeLightDarkIcon from 'mdi-react/ThemeLightDarkIcon'
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
        <button className="theme-switcher__button btn btn-icon">
            <ThemeLightDarkIcon className="icon-inline" />
        </button>
    </div>
)
