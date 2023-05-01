import { VSCodeProgressRing } from '@vscode/webview-ui-toolkit/react'

export const LoadingPage: React.FunctionComponent<{}> = () => (
    <div className="outer-container">
        <div className="inner-container">
            <div className="non-transcript-container">
                <VSCodeProgressRing />
            </div>
        </div>
    </div>
)
