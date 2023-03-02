/* eslint-disable @typescript-eslint/explicit-member-accessibility */
/* eslint-disable @typescript-eslint/explicit-function-return-type */
/* eslint-disable no-undef */
const MAX_HEIGHT = 192

let tabsController = null
let askController = null
const debugLog = []

class ChatController {
	constructor(containerId, recipeContainerId, vscode, gettingStartedHTML) {
		this.containerId = containerId
		this.recipeContainerId = recipeContainerId
		this.vscode = vscode
		this.gettingStartedHTML = gettingStartedHTML

		this.containerElement = document.getElementById(containerId)
		this.inputElement = this.containerElement.querySelector('.chat-input')
		this.submitElement = this.containerElement.querySelector('.submit-container')
		this.resetElement = this.containerElement.querySelector('.reset-conversation')
		this.transcriptContainerElement = this.containerElement.querySelector('.transcript-container')

		this.initListeners()
	}

	initListeners() {
		// TODO(beyang): handle destruction
		this.inputElement.addEventListener('keydown', e => {
			if (e.key === 'Enter' && !e.shiftKey) {
				if (e.target.value.trim().length === 0) {
					return
				}
				this.vscode.postMessage({ command: 'submit', text: e.target.value })
				e.target.value = ''
				e.preventDefault()

				setTimeout(() => this.resizeInput(), 0)
			}
		})
		this.inputElement.addEventListener('input', () => this.resizeInput())

		this.submitElement.addEventListener('click', () => {
			if (this.inputElement.value.trim().length === 0) {
				return
			}
			this.vscode.postMessage({ command: 'submit', text: this.inputElement.value })
			this.inputElement.value = ''
		})

		this.resetElement.addEventListener('click', () => {
			this.vscode.postMessage({ command: 'reset' })
		})

		if (this.recipeContainerId) {
			const recipeContainer = document.getElementById(this.recipeContainerId)
			const onRecipeButtonClick = e => {
				e.preventDefault()
				this.vscode.postMessage({
					command: 'executeRecipe',
					recipe: e.target.dataset.recipe,
				})
			}
			const recipeButtons = recipeContainer.querySelectorAll('.btn-recipe')
			recipeButtons.forEach(button => button.addEventListener('click', onRecipeButtonClick))
		}

		// Transcript click listener
		this.containerElement.addEventListener('click', event => {
			if (event.target.dataset.contextfiles !== undefined) {
				this.handleToggleContextFilesExpand(event)
			} else if (event.target.dataset.feedbacksentiment !== undefined) {
				this.handleFeedback(event)
			}
		})
	}

	handleToggleContextFilesExpand(event) {
		const contextFilesStr = event.target.dataset.contextfiles
		if (!contextFilesStr) {
			return
		}
		const contextFiles = JSON.parse(contextFilesStr)

		const isExpanded = JSON.parse(event.target.dataset.isexpanded || 'false')
		const willExpand = !isExpanded
		if (willExpand) {
			event.target.innerHTML = getContextFilesString(contextFiles, true)
		} else {
			event.target.innerHTML = getContextFilesString(contextFiles, false)
		}
		event.target.dataset.isexpanded = `${willExpand}`
	}

	handleFeedback(event) {
		const sentiment = event.target.dataset.feedbacksentiment

		// Update UI
		let node = event.target
		while (node) {
			if (node.classList.contains('feedback-container')) {
				break
			}
			node = node.parentElement
		}
		node.innerHTML = 'Response recorded'
		this.vscode.postMessage({
			command: 'feedback',
			feedback: {
				sentiment,
			},
		})
	}

	resizeInput() {
		this.inputElement.style.height = 0
		const height = Math.min(MAX_HEIGHT, this.inputElement.scrollHeight)
		this.inputElement.style.height = `${height}px`
		this.inputElement.style.overflowY = height >= MAX_HEIGHT ? 'auto' : 'hidden'
	}

	renderMessages(messages, messageInProgress) {
		const visibleMessages = messages.filter(message => !message.hidden)
		const messageElements = visibleMessages.map((message, idx) =>
			getMessageBubble(
				message.speaker,
				message.displayText,
				message.timestamp,
				message.contextFiles,
				idx === visibleMessages.length - 1
			)
		)

		const messageInProgressElement = messageInProgress
			? getMessageInProgressBubble(
					messageInProgress.speaker,
					messageInProgress.displayText,
					messageInProgress.contextFiles
			  )
			: ''

		if (messageInProgress) {
			this.inputElement.setAttribute('disabled', '')
			this.submitElement.style.cursor = 'default'
		} else {
			this.inputElement.removeAttribute('disabled')
			this.submitElement.style.cursor = 'pointer'
		}

		if ((!messages || messages.length === 0) && !messageInProgress && this.gettingStartedHTML) {
			this.transcriptContainerElement.innerHTML = this.gettingStartedHTML
		} else {
			this.transcriptContainerElement.innerHTML = messageElements.join('') + messageInProgressElement
		}

		setTimeout(() => {
			if (messageInProgress && messageInProgress.displayText.length === 0) {
				this.containerElement.querySelector('.bubble-loader')?.scrollIntoView()
			} else if (!messageInProgress && messages.length > 0) {
				this.containerElement.querySelector('.bubble-row:last-child')?.scrollIntoView()
			}
		}, 0)
	}
}

class TOSController {
	constructor(containerId, viewTOSId, vscode) {
		this.containerElement = document.getElementById(containerId)
		this.viewTOSElement = document.getElementById(viewTOSId)
		this.vscode = vscode

		this.initListeners()
		this.updateUI()
	}

	getTOSVersion() {
		return parseInt(this.containerElement.dataset.tosversion)
	}

	getAcceptedTOSVersion() {
		return parseInt(this.containerElement.dataset.tosacceptedversion)
	}

	updateUI() {
		if (isNaN(this.getAcceptedTOSVersion())) {
			console.error('acceptedTOSVersion was NaN')
			return
		}
		if (this.getAcceptedTOSVersion() === this.getTOSVersion()) {
			this.containerElement.classList.add('tos-container-accepted')
		}
	}

	initListeners() {
		this.containerElement.querySelector('#tos-accept').addEventListener('click', event => {
			event.preventDefault()
			this.containerElement.dataset.tosacceptedversion = this.containerElement.dataset.tosversion
			this.vscode.postMessage({ command: 'acceptTOS', version: parseInt(this.containerElement.dataset.tosversion) })
			this.updateUI()
		})
		this.viewTOSElement.addEventListener('click', event => {
			event.preventDefault()
			this.containerElement.classList.remove('tos-container-accepted')
		})
	}
}

function onInitialize() {
	const vscode = acquireVsCodeApi()
	tabsController = new TabsController({ selectedTab: 'ask' })
	askController = new ChatController('container-ask', 'container-recipes', vscode, gettingStartedHTML)
	tosController = new TOSController('tos-container', 'view-tos', vscode)
	vscode.postMessage({ command: 'initialized', containerId: 'container-ask' })
}

function onMessage(event) {
	switch (event.data.type) {
		case 'transcript':
			askController.renderMessages(event.data.messages, event.data.messageInProgress)
			break
		case 'showTab':
			if (tabsController) {
				tabsController.setSelectedTab(event.data.tab)
			}
			break
		case 'debug':
			debugLog.push(event.data.message)
			renderDebugLog(debugLog)
			break
	}
}

const gettingStartedHTML = `
<div class="container-getting-started">
	<h1>Example questions</h1>
	<ul>
		<li>What are the most popular Go CLI libraries?</li>
		<li>Write a function that parses JSON in Python.</li>
		<li>Summarize the code in this file.</li>
		<li>Which files handle SAML authentication in my codebase?</li>
	</ul>

	<h1>Recommendations</h1>
	<ul>
		<li>Visit the <strong>Recipes</strong> tab for special actions like <strong>Write a unit test</strong> or <strong>Summarize code history</strong>.</li>
		<li>
			Cody tells you which files it reads to respond to your message. If this list of files looks wrong, try copying the relevant code (up to 20KB) into your message like this:<br/>
			<blockquote>
			\`\`\`<br/>
			{code}</br/>
			\`\`\`</br/>
			Explain the code above (or whatever your question is).
			</blockquote>
		</li>
		<li>
			Use the
			<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" style="fill:#bbb">
				<path d="M12 16c1.671 0 3-1.331 3-3s-1.329-3-3-3-3 1.331-3 3 1.329 3 3 3z"></path>
				<path d="M20.817 11.186a8.94 8.94 0 0 0-1.355-3.219 9.053 9.053 0 0 0-2.43-2.43 8.95 8.95 0 0 0-3.219-1.355 9.028 9.028 0 0 0-1.838-.18V2L8 5l3.975 3V6.002c.484-.002.968.044 1.435.14a6.961 6.961 0 0 1 2.502 1.053 7.005 7.005 0 0 1 1.892 1.892A6.967 6.967 0 0 1 19 13a7.032 7.032 0 0 1-.55 2.725 7.11 7.11 0 0 1-.644 1.188 7.2 7.2 0 0 1-.858 1.039 7.028 7.028 0 0 1-3.536 1.907 7.13 7.13 0 0 1-2.822 0 6.961 6.961 0 0 1-2.503-1.054 7.002 7.002 0 0 1-1.89-1.89A6.996 6.996 0 0 1 5 13H3a9.02 9.02 0 0 0 1.539 5.034 9.096 9.096 0 0 0 2.428 2.428A8.95 8.95 0 0 0 12 22a9.09 9.09 0 0 0 1.814-.183 9.014 9.014 0 0 0 3.218-1.355 8.886 8.886 0 0 0 1.331-1.099 9.228 9.228 0 0 0 1.1-1.332A8.952 8.952 0 0 0 21 13a9.09 9.09 0 0 0-.183-1.814z"></path>
			</svg>
			button in the upper right to reset the chat when you want to start a new line of thought. Cody does not remember anything outside the current chat.
		</li>
		<li>Use the feedback buttons when Cody messes up. We will use your feedback to make Cody better.</li>
	</ul>
</div>
`

const messageBubbleTemplate = `
<div class="bubble-row {type}-bubble-row">
	<div class="bubble {type}-bubble">
		<div class="bubble-content {type}-bubble-content">
			{contextFiles}
			{text}
			{feedback}
		</div>
		<div class="bubble-footer {type}-bubble-footer">
			{footer}
		</div>
	</div>
</div>
`

function getContextFilesString(contextFiles, expand) {
	contextFiles = contextFiles.map(f => f.replace(/^\.\//, ''))
	if (expand) {
		return `Cody read:\n<ul>${contextFiles.map(f => '<li>' + f + '</li>').join('\n')}</ul>`
	}
	return `Cody read ${contextFiles[0]} and ${contextFiles.length - 1} other files`
}

function getContextFilesHTML(contextFiles, expand) {
	if (!contextFiles || contextFiles.length === 0) {
		return ''
	}
	if (contextFiles.length === 1) {
		return `<span style="font-style: italic;">Cody read ${contextFiles[0]}</span>`
	}
	return `<p data-contextfiles='${JSON.stringify(contextFiles)}' style="font-style: italic;">${getContextFilesString(
		contextFiles,
		expand
	)}</p>`
}

function getMessageBubble(author, text, timestamp, contextFiles, includeFeedback) {
	const bubbleType = author === 'bot' ? 'bot' : 'human'
	const authorName = author === 'bot' ? 'Cody' : 'Me'
	return messageBubbleTemplate
		.replace(/{type}/g, bubbleType)
		.replace('{text}', text)
		.replace(
			'{feedback}',
			includeFeedback && author === 'bot'
				? '<div class="feedback-container" style="font-style: italic;">Was this response helpful? <button data-feedbacksentiment="good" class="btn-feedback">&#128077;</button> <button data-feedbacksentiment="bad" class="btn-feedback">&#128078;</button></div>'
				: ''
		)
		.replace('{contextFiles}', getContextFilesHTML(contextFiles, false))
		.replace('{footer}', timestamp ? `${authorName} &middot; ${timestamp}` : `<i>${authorName} is writing...</i>`)
}

function getMessageInProgressBubble(author, text, contextFiles) {
	if (text.length === 0) {
		const loader = `
		<div class="bubble-loader">
			<div class="bubble-loader-dot"></div>
			<div class="bubble-loader-dot"></div>
			<div class="bubble-loader-dot"></div>
		</div>`
		return getMessageBubble(author, loader, null, null, false)
	}
	return getMessageBubble(author, text, null, contextFiles, false)
}

const debugMessageTemplate = `
<div class="debug-message">
	<pre>{message6fc87d4}</pre>
</div>
`
function renderDebugLog(debugMessages) {
	const debugContainerElement = document.querySelector('.debug-container')
	if (!debugContainerElement) {
		return
	}

	const escapeEl = document.createElement('textarea')
	debugContainerElement.innerHTML = debugMessages
		.map(message => {
			escapeEl.textContent = message
			return debugMessageTemplate.replace('{message6fc87d4}', escapeEl.innerHTML)
		})
		.join('\n')
}

window.addEventListener('message', onMessage)
document.addEventListener('DOMContentLoaded', onInitialize)
