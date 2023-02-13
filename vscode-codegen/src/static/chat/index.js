const MAX_HEIGHT = 192

let tabsController = null
let debugLog = []

// TODO: We need a design for the chat empty state.
function onInitialize() {
	const vscode = acquireVsCodeApi()
	const inputElement = document.getElementById('input')
	const submitElement = document.querySelector('.submit-container')
	const resetElement = document.querySelector('.reset-conversation')

	const resizeInput = () => {
		inputElement.style.height = 0
		const height = Math.min(MAX_HEIGHT, inputElement.scrollHeight)
		inputElement.style.height = `${height}px`
		inputElement.style.overflowY = height >= MAX_HEIGHT ? 'auto' : 'hidden'
	}

	tabsController = new TabsController({ selectedTab: 'chat' })

	inputElement.addEventListener('keydown', e => {
		if (e.key === 'Enter' && !e.shiftKey) {
			if (e.target.value.trim().length === 0) {
				return
			}
			vscode.postMessage({ command: 'submit', text: e.target.value })
			e.target.value = ''
			e.preventDefault()

			setTimeout(resizeInput, 0)
		}
	})

	inputElement.addEventListener('input', resizeInput)

	submitElement.addEventListener('click', () => {
		if (inputElement.value.trim().length === 0) {
			return
		}
		vscode.postMessage({ command: 'submit', text: inputElement.value })
		inputElement.value = ''
	})

	const onRecipeButtonClick = e => {
		vscode.postMessage({ command: 'executeRecipe', recipe: e.target.dataset.recipe })
	}
	const recipeButtons = document.querySelectorAll('.btn-recipe')
	recipeButtons.forEach(button => button.addEventListener('click', onRecipeButtonClick))

	resetElement.addEventListener('click', () => {
		vscode.postMessage({ command: 'reset' })
	})

	vscode.postMessage({ command: 'initialized' })
}

function onMessage(event) {
	switch (event.data.type) {
		case 'transcript':
			renderMessages(event.data.messages, event.data.messageInProgress)
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

const messageBubbleTemplate = `
<div class="bubble-row {type}-bubble-row">
	<div class="bubble {type}-bubble">
		<div class="bubble-content {type}-bubble-content">{text}</div>
		<div class="bubble-footer {type}-bubble-footer">
			{footer}
		</div>
	</div>
</div>
`

function getMessageBubble(author, text, timestamp) {
	const bubbleType = author === 'bot' ? 'bot' : 'human'
	const authorName = author === 'bot' ? 'Cody' : 'Me'
	return messageBubbleTemplate
		.replace(/{type}/g, bubbleType)
		.replace('{text}', text)
		.replace('{footer}', timestamp ? `${authorName} &middot; ${timestamp}` : `<i>${authorName} is writing...</i>`)
}

function getMessageInProgressBubble(author, text) {
	if (text.length === 0) {
		const loader = `
		<div class="bubble-loader">
			<div class="bubble-loader-dot"></div>
			<div class="bubble-loader-dot"></div>
			<div class="bubble-loader-dot"></div>
		</div>`
		return getMessageBubble(author, loader, null)
	}
	return getMessageBubble(author, text, null)
}

function renderMessages(messages, messageInProgress) {
	const inputElement = document.getElementById('input')
	const submitElement = document.querySelector('.submit-container')
	const transcriptContainerElement = document.querySelector('.transcript-container')

	const messageElements = messages
		.filter(message => !message.hidden)
		.map(message => getMessageBubble(message.speaker, message.displayText, message.timestamp))

	const messageInProgressElement = messageInProgress
		? getMessageInProgressBubble(messageInProgress.speaker, messageInProgress.displayText)
		: ''

	if (messageInProgress) {
		inputElement.setAttribute('disabled', '')
		submitElement.style.cursor = 'default'
	} else {
		inputElement.removeAttribute('disabled')
		submitElement.style.cursor = 'pointer'
	}

	transcriptContainerElement.innerHTML = messageElements.join('') + messageInProgressElement

	setTimeout(() => {
		if (messageInProgress && messageInProgress.displayText.length === 0) {
			document.querySelector('.bubble-loader')?.scrollIntoView()
		} else if (!messageInProgress && messages.length > 0) {
			document.querySelector('.bubble-row:last-child')?.scrollIntoView()
		}
	}, 0)
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
