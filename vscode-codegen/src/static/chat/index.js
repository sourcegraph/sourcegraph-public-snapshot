const INPUT_HEIGHT = 24;

// TODO: We need a design for the chat empty state.
function onInitialize() {
	const vscode = acquireVsCodeApi();
	const inputElement = document.getElementById("input");
	const submitElement = document.querySelector(".submit-container");
	const resetElement = document.querySelector(".reset-conversation");

	inputElement.addEventListener("keydown", (e) => {
		if (e.key === "Enter" && !e.shiftKey) {
			if (e.target.value.trim().length === 0) {
				return;
			}
			vscode.postMessage({ command: "submit", text: e.target.value });
			e.target.value = "";
			e.preventDefault();
		}

		// Resize the input on each added line up to a maximum height.
		// TODO: This does not handle long overflowing lines.
		setTimeout(() => {
			const lines = inputElement.value.split("\n").length;
			input.parentNode.style.height =
				Math.min(8 * INPUT_HEIGHT, lines * INPUT_HEIGHT) + "px";
		}, 0);
	});

	submitElement.addEventListener("click", () => {
		if (inputElement.value.trim().length === 0) {
			return;
		}
		vscode.postMessage({ command: "submit", text: inputElement.value });
		inputElement.value = "";
	});

	resetElement.addEventListener("click", () => {
		vscode.postMessage({ command: "reset" });
	});

	vscode.postMessage({ command: "initialized" });
}

function onMessage(event) {
	switch (event.data.type) {
		case "transcript":
			renderMessages(event.data.messages, event.data.messageInProgress);
			break;
	}
}

const messageBubbleTemplate = `
<div class="bubble-row {type}-bubble-row">
	<div class="bubble {type}-bubble">
		<div class="bubble-content {type}-bubble-content">{text}</div>
		<div class="bubble-footer {type}-bubble-footer">
			{author} &middot; {timestamp}
		</div>
	</div>
</div>
`;

function getMessageBubble(author, text, timestamp) {
	const bubbleType = author === "bot" ? "bot" : "human";
	const authorName = author === "bot" ? "Cody" : "Me";
	return messageBubbleTemplate
		.replace(/{type}/g, bubbleType)
		.replace("{author}", authorName)
		.replace("{text}", text)
		.replace("{timestamp}", timestamp);
}

function getMessageInProgressBubble(author, text, timestamp) {
	if (text.length === 0) {
		const loader = `
		<div class="bubble-loader">
			<div class="bubble-loader-dot"></div>
			<div class="bubble-loader-dot"></div>
			<div class="bubble-loader-dot"></div>
		</div>`;
		return getMessageBubble(author, loader, timestamp);
	}
	return getMessageBubble(author, `${text}█`, timestamp);
}

function renderMessages(messages, messageInProgress) {
	const inputElement = document.getElementById("input");
	const submitElement = document.querySelector(".submit-container");
	const transcriptContainerElement = document.querySelector(
		".transcript-container"
	);

	const messageElements = messages
		.filter((message) => !message.hidden)
		.map((message) =>
			getMessageBubble(message.speaker, message.displayText, message.timestamp)
		);

	const messageInProgressElement = messageInProgress
		? getMessageInProgressBubble(
				messageInProgress.speaker,
				messageInProgress.displayText,
				messageInProgress.timestamp
		  )
		: "";

	if (messageInProgress) {
		inputElement.setAttribute("disabled", "");
		submitElement.style.cursor = "default";
	} else {
		inputElement.removeAttribute("disabled");
		submitElement.style.cursor = "pointer";
	}

	transcriptContainerElement.innerHTML =
		messageElements.join("") + messageInProgressElement;

	setTimeout(() => {
		if (messageInProgress && messageInProgress.displayText.length === 0) {
			document.querySelector(".bubble-loader")?.scrollIntoView();
		} else if (!messageInProgress && messages.length > 0) {
			document.querySelector(".bubble-row:last-child")?.scrollIntoView();
		}
	}, 0);
}

window.addEventListener("message", onMessage);
document.addEventListener("DOMContentLoaded", onInitialize);
