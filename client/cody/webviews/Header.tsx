import { VSCodeTag } from '@vscode/webview-ui-toolkit/react'

import './Header.css'

import { CodyColoredSvg } from '@sourcegraph/cody-ui/src/utils/icons'

export const Header: React.FunctionComponent = () => (
    <div className="header-container">
        <div className="header-container-left">
            <div className="header-logo">
                <CodyColoredSvg />
            </div>
            <div className="header-title">
                <span className="header-cody">Cody</span>
                <VSCodeTag className="tag-beta">experimental</VSCodeTag>
            </div>
        </div>
        <div className="header-container-right" />
    </div>
)
