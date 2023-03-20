import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { ContextFiles } from './Chat'

import './UserHistory.css'
import './Chat.css'

import { ChatHistory, ChatMessage, View } from './utils/types'

interface HistoryProps {
    setView: (view: View) => void
    userHistory: ChatHistory | null
}

export const UserHistory: React.FunctionComponent<React.PropsWithChildren<HistoryProps>> = ({
    setView,
    userHistory,
}) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="bubble-content">
                {userHistory &&
                    [...Object.entries(userHistory)].map(chat => (
                        <div className="history-container">
                            <VSCodeButton className="history-title-container" type="button">
                                {chat[0]}
                            </VSCodeButton>
                            <div className="history-convo-container">
                                {chat[1].map((message: ChatMessage, index: number) => (
                                    <div key={index} className="history-bubble-container">
                                        {message.displayText && (
                                            <p dangerouslySetInnerHTML={{ __html: message.displayText }} />
                                        )}
                                        {message.contextFiles && message.contextFiles.length > 0 && (
                                            <ContextFiles contextFiles={message.contextFiles} />
                                        )}
                                    </div>
                                ))}
                            </div>
                        </div>
                    ))}
            </div>
        </div>
    </div>
)
