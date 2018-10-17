import { SettingsOutlineIcon } from 'mdi-react'
import * as React from 'react'

export interface OptionsHeaderProps {
    className?: string
    version: string
    onSettingsClick: (event: React.MouseEvent<HTMLButtonElement>) => void
    assetsDir?: string
}

export const OptionsHeader: React.SFC<OptionsHeaderProps> = ({
    className,
    version,
    assetsDir,
    onSettingsClick,
}: OptionsHeaderProps) => (
    <div className={`options-header ${className || ''}`}>
        <img src={`${assetsDir || ''}/img/sourcegraph-logo.svg`} className="options-header__logo" />
        <div className="options-header__right">
            <span>v{version}</span>
            <button className="options-header__right__settings btn btn-icon" onClick={onSettingsClick}>
                <SettingsOutlineIcon className="icon-inline" />
            </button>
        </div>
    </div>
)
