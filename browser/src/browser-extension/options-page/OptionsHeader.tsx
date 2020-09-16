import SettingsOutlineIcon from 'mdi-react/SettingsOutlineIcon'
import { Toggle } from '../../../../shared/src/components/Toggle'
import * as React from 'react'

export interface OptionsHeaderProps {
    className?: string
    version: string
    assetsDir?: string
    isActivated: boolean
    onClickExpandOptionsMenu: (event: React.MouseEvent<HTMLButtonElement>) => void
    onToggleActivationClick: (value: boolean) => void
}

export const OptionsHeader: React.FunctionComponent<OptionsHeaderProps> = ({
    className,
    version,
    assetsDir,
    isActivated,
    onClickExpandOptionsMenu,
    onToggleActivationClick,
}: OptionsHeaderProps) => (
    <div className={`options-header ${className || ''}`}>
        <div>
            <img src={`${assetsDir || ''}/img/sourcegraph-logo.svg`} className="options-header__logo" />
            <div className="options-header__version">v{version}</div>
        </div>
        <div className="options-header__right">
            <button type="button" className="options-header__settings btn btn-icon" onClick={onClickExpandOptionsMenu}>
                <SettingsOutlineIcon className="icon-inline" />
            </button>
            <Toggle
                value={isActivated}
                onToggle={onToggleActivationClick}
                title={isActivated ? 'Toggle to disable extension' : 'Toggle to enable extension'}
            />
        </div>
    </div>
)
