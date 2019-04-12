import SettingsOutlineIcon from 'mdi-react/SettingsOutlineIcon'
import * as React from 'react'

export interface OptionsHeaderProps {
    className?: string
    version: string
    assetsDir?: string
    onSettingsClick: (event: React.MouseEvent<HTMLButtonElement>) => void
}

export const OptionsHeader: React.FunctionComponent<OptionsHeaderProps> = ({
    className,
    version,
    assetsDir,
    onSettingsClick,
}: OptionsHeaderProps) => (
    <div className={`options-header ${className || ''}`}>
        <img src={`${assetsDir || ''}/img/sourcegraph-logo.svg`} className="options-header__logo" />
        <div className="options-header__right">
            <span>v{version}</span>
            <button className="options-header__settings btn btn-icon" onClick={onSettingsClick}>
                <SettingsOutlineIcon className="icon-inline" />
            </button>
        </div>
    </div>
)
