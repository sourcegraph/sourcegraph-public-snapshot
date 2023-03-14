import { VSCodeButton, VSCodeTag } from '@vscode/webview-ui-toolkit/react'

import { CodySvg, ResetSvg } from './utils/icons'

import './Header.css'

interface HeaderProps {
    showResetButton: boolean
    onResetClick: () => void
}

export const Header: React.FunctionComponent<React.PropsWithChildren<HeaderProps>> = ({
    showResetButton,
    onResetClick,
}) => (
    <div className="header-container">
        <div className="header-container-left">
            <div className="header-logo">
                <CodySvg />
            </div>
            <div className="header-title">
                <span className="header-cody">Cody</span>
                <VSCodeTag className="tag-beta">Beta</VSCodeTag>
            </div>
        </div>
        <div className="header-container-right">
            <VSCodeButton
                className="reset-conversation"
                title="Start a new conversation with Cody"
                onClick={() => onResetClick()}
                appearance="icon"
                type="button"
            >
                {showResetButton && <ResetSvg />}
            </VSCodeButton>
        </div>
    </div>
)
