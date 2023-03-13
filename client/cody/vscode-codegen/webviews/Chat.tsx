import React, { useCallback, useState } from 'react'

import { VSCodeButton, VSCodeTextArea } from '@vscode/webview-ui-toolkit/react'

import { Tips } from './Tips'
import { DownArrowSvg, RightArrowSvg, SubmitSvg } from './utils/icons'
import { ChatMessage } from './utils/types'
import { WebviewMessage, vscodeAPI } from './utils/vscodeAPI'

import './Chat.css'

interface ChatboxProps {
	messageInProgress: ChatMessage | null
	setMessageInProgress: (transcript: ChatMessage | null) => void
	transcript: ChatMessage[]
	setTranscript: (transcripts: ChatMessage[]) => void
	devMode: boolean
}

export const Chat: React.FunctionComponent<React.PropsWithChildren<ChatboxProps>> = ({
	messageInProgress,
	setMessageInProgress,
	transcript,
	setTranscript,
	devMode,
}) => {
	const [inputRows, setInputRows] = useState(5)
	const [formInput, setFormInput] = useState('')

	const inputHandler = useCallback(
		(inputValue: string) => {
			const rowsCount = inputValue.match(/\n/g)?.length
			if (rowsCount) {
				setInputRows(rowsCount < 5 ? 5 : rowsCount > 25 ? 25 : rowsCount)
			} else {
				setInputRows(5)
			}
			setFormInput(inputValue)
		},
		[setFormInput]
	)

	const onChatKeyDown = async (event: React.KeyboardEvent<HTMLDivElement>): Promise<void> => {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault()
			event.stopPropagation()
			await onChatSubmit()
		}
	}

	const onChatSubmit = useCallback(async () => {
		setInputRows(5)
		const chatMsg: ChatMessage = { speaker: 'you', displayText: formInput, timestamp: getShortTimestamp() }
		setMessageInProgress({ speaker: 'bot', displayText: '', timestamp: getShortTimestamp() })
		setTranscript([...transcript, chatMsg])

		vscodeAPI.postMessage({ command: 'submit', text: formInput } as WebviewMessage)

		if (formInput === '/reset') {
			setMessageInProgress(null)
		}
		setFormInput('')
	}, [formInput, setTranscript, setMessageInProgress, transcript])

	const bubbleClassName = (speaker: string): string => (speaker === 'you' ? 'human' : 'bot')

	return (
		<div className="inner-container">
			<div className={`${transcript.length >= 1 ? '' : 'non-'}transcript-container`}>
				{/* Show Tips view if no conversation has happened */}
				{transcript.length === 0 && !messageInProgress && <Tips />}
				{transcript.length > 0 && (
					<div className="bubble-container">
						{transcript.map((message, index) => (
							<div
								key={message.timestamp}
								className={`bubble-row ${bubbleClassName(message.speaker)}-bubble-row`}
							>
								<div className={`bubble ${bubbleClassName(message.speaker)}-bubble`}>
									<div
										className={`bubble-content ${bubbleClassName(message.speaker)}-bubble-content`}
									>
										{message.displayText && (
											<p dangerouslySetInnerHTML={{ __html: message.displayText }} />
										)}
										{message.contextFiles && message.contextFiles.length > 0 && (
											<ContextFiles contextFiles={message.contextFiles} />
										)}
									</div>
									<div className={`bubble-footer ${bubbleClassName(message.speaker)}-bubble-footer`}>
										<span className="bubble-footer-timestamp">{`${
											message.speaker === 'bot' ? 'Cody' : 'Me'
										} · ${message.timestamp}`}</span>
										{/* Only show feedback for the last message. */}
										{message.speaker === 'bot' && index === transcript.length - 1 && devMode && (
											<FeedbackContainer />
										)}
									</div>
								</div>
							</div>
						))}

						{messageInProgress && messageInProgress.speaker === 'bot' && (
							<div className="bubble-row bot-bubble-row">
								<div className="bubble bot-bubble">
									<div className="bubble-content bot-bubble-content">
										{messageInProgress.displayText ? (
											<p dangerouslySetInnerHTML={{ __html: messageInProgress.displayText }} />
										) : (
											<div className="bubble-loader">
												<div className="bubble-loader-dot" />
												<div className="bubble-loader-dot" />
												<div className="bubble-loader-dot" />
											</div>
										)}
									</div>
									<div className="bubble-footer bot-bubble-footer">
										<span>Cody is typing...</span>
									</div>
								</div>
							</div>
						)}
					</div>
				)}
			</div>
			<form className="input-row">
				<VSCodeTextArea
					className="chat-input"
					rows={inputRows}
					name="user-query"
					value={formInput}
					autofocus={true}
					disabled={!!messageInProgress}
					required={true}
					onInput={({ target }) => {
						const { value } = target as HTMLInputElement
						inputHandler(value)
					}}
					onKeyDown={onChatKeyDown}
				/>
				<VSCodeButton className="submit-button" appearance="icon" type="button" onClick={onChatSubmit}>
					<SubmitSvg />
				</VSCodeButton>
			</form>
		</div>
	)
}

const ContextFiles: React.FunctionComponent<{ contextFiles: string[] }> = ({ contextFiles }) => {
	const [isExpanded, setIsExpanded] = useState(false)

	if (contextFiles.length === 1) {
		return (
			<p>
				Cody read <code className="context-file">{contextFiles[0]}</code> file to provide an answer.
			</p>
		)
	}

	if (isExpanded) {
		return (
			<p className="context-files-expanded">
				<span className="context-files-toggle-icon" onClick={() => setIsExpanded(false)}>
					<DownArrowSvg />
				</span>
				<div>
					<div className="context-files-list-title" onClick={() => setIsExpanded(false)}>
						Cody read the following files to provide an answer:
					</div>
					<ul>
						{contextFiles.map(file => (
							<li key={file}>
								<code className="context-file">{file}</code>
							</li>
						))}
					</ul>
				</div>
			</p>
		)
	}

	return (
		<p className="context-files-collapsed" onClick={() => setIsExpanded(true)}>
			<span className="context-files-toggle-icon">
				<RightArrowSvg />
			</span>
			<div className="context-files-collapsed-text">
				<span>
					Cody read <code className="context-file">{contextFiles[0]}</code> and {contextFiles.length - 1}{' '}
					other {contextFiles.length > 2 ? 'files' : 'file'} to provide an answer.
				</span>
			</div>
		</p>
	)
}

const FeedbackContainer = React.memo(() => (
	<div className="feedback-container">
		<div className="feedback-container-title">Was the response helpful?</div>
		<div className="feedback-container-emojis">
			<VSCodeButton data-feedbacksentiment="good" className="feedback-button">
				&#128077;
			</VSCodeButton>{' '}
			<VSCodeButton data-feedbacksentiment="bad" className="feedback-button">
				&#128078;
			</VSCodeButton>
		</div>
	</div>
))

export function getShortTimestamp(): string {
	const date = new Date()
	return `${padTimePart(date.getHours())}:${padTimePart(date.getMinutes())}`
}

function padTimePart(timePart: number): string {
	return timePart < 10 ? `0${timePart}` : timePart.toString()
}
