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
			this.vscode.postMessage({ command: 'submit', text: inputElement.value })
			this.inputElement.value = ''
		})

		this.resetElement.addEventListener('click', () => {
			this.vscode.postMessage({ command: 'reset' })
		})

		if (this.recipeContainerId) {
			const recipeContainer = document.getElementById(this.recipeContainerId)
			const onRecipeButtonClick = e => {
				this.vscode.postMessage({
					command: 'executeRecipe',
					recipe: e.target.dataset.recipe,
				})
			}
			const recipeButtons = recipeContainer.querySelectorAll('.btn-recipe')
			recipeButtons.forEach(button => button.addEventListener('click', onRecipeButtonClick))
		}
	}

	resizeInput() {
		this.inputElement.style.height = 0
		const height = Math.min(MAX_HEIGHT, this.inputElement.scrollHeight)
		this.inputElement.style.height = `${height}px`
		this.inputElement.style.overflowY = height >= MAX_HEIGHT ? 'auto' : 'hidden'
	}

	renderMessages(messages, messageInProgress) {
		const messageElements = messages
			.filter(message => !message.hidden)
			.map(message => getMessageBubble(message.speaker, message.displayText, message.timestamp, message.contextFiles))

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

function onInitialize() {
	const vscode = acquireVsCodeApi()
	tabsController = new TabsController({ selectedTab: 'ask' })
	askController = new ChatController('container-ask', 'container-recipes', vscode, gettingStartedHTML)
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
	<h1>Tips and tricks</h1>
	<p>Here are some examples of what you can ask Cody:</p>
	<ul>
		<li>What are the most popular Go CLI libraries?</li>
		<li>Write a function that parses JSON in Python</li>
		<li>What changed in my codebase in the last day?</li>
		<li>Summarize the code in this file.</li>
		<li>What's wrong with this code?</li>
	</ul>

	<p>Recommendations:</p>
	<ul>
		<li>Make your questions detailed and specific. The more context Cody has, the better Cody's responses will be.</li>
	</ul>
</div>
`

const messageBubbleTemplate = `
<div class="bubble-row {type}-bubble-row">
	<div class="bubble {type}-bubble">
		<div class="bubble-content {type}-bubble-content">
			{contextFiles}
			{text}
		</div>
		<div class="bubble-footer {type}-bubble-footer">
			{footer}
		</div>
	</div>
</div>
`

function toggleContextFilesExpand(event) {
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

function getMessageBubble(author, text, timestamp, contextFiles) {
	const bubbleType = author === 'bot' ? 'bot' : 'human'
	const authorName = author === 'bot' ? 'Cody' : 'Me'
	return messageBubbleTemplate
		.replace(/{type}/g, bubbleType)
		.replace('{text}', text)
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
		return getMessageBubble(author, loader, null)
	}
	return getMessageBubble(author, text, null, contextFiles)
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
document.addEventListener('click', toggleContextFilesExpand)
