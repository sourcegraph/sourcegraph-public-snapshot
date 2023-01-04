import * as vscode from "vscode";
import {
	Message,
	WSChatRequest,
	WSChatResponse,
} from "@sourcegraph/cody-common";
import { WSClient } from "./wsclient";

export class ChatViewProvider implements vscode.WebviewViewProvider {
	messages: Message[] = [
		{
			speaker: "bot",
			text: "How can I help?",
		},
	];
	messageInProgress?: Message;
	inputContents = "";

	constructor(private wsclient: WSChatClient) {}

	resolveWebviewView(
		webviewView: vscode.WebviewView,
		context: vscode.WebviewViewResolveContext<unknown>,
		token: vscode.CancellationToken
	): void | Thenable<void> {
		webviewView.webview.html = this.renderView();
		webviewView.webview.options = {
			enableScripts: true, // TODO(beyang): temporary, switch to non-inline scripts
			portMapping: [
				// TODO(beyang): define portmapping, so that chat view can be completely self-contained
			],
		};
		const updateView = () => {
			webviewView.webview.html = this.renderView();
		};

		webviewView.webview.onDidReceiveMessage((message) => {
			switch (message.command) {
				case "typed":
					this.inputContents = message.text;
					break;
				case "submit":
					if (this.messageInProgress) {
						break;
					}
					this.inputContents = "";
					this.messageInProgress = {
						speaker: "bot",
						text: "",
					};

					this.messages.push({
						speaker: "you",
						text: message.text,
					});
					updateView();

					this.wsclient.chat(this.messages, {
						onChange: (text) => {
							this.messageInProgress = {
								speaker: "bot",
								text,
							};
							updateView();
						},
						onComplete: (text) => {
							this.messageInProgress = undefined;
							this.messages.push({
								speaker: "bot",
								text,
							});
							updateView();
						},
						onError: (err) => {
							vscode.window.showErrorMessage(err);
						},
					});

					// TODO(beyang): prevent user from submitting new request while existing one pending
					break;
			}
		});
	}

	renderView(): string {
		return `<!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Cody</title>
    </head>
    <body>
    <style>
        html, body {
            height: 100%;
            margin: 0;
            padding: 0;
        }
        .container {
            height: 100%;
            display: flex;
            flex-direction: column;
        }
        .transcript-container {
            flex: 2 1 100%;
        }
        .input-container {
            flex: none;
            padding: 1rem;
            display: flex;
        }
        .input-container form {
            flex: 1;
            display: flex;
        }
        .input {
            flex: 1;
        }
        .message-container {
            padding: 0 1rem 0 1rem;
            margin-bottom: 1rem;
        }
        .message-container:last {
            margin-bottom: 0;
        }
        .message-speaker {
            font-weight: bold;
        }
    </style>
    <div class="container">
        <div class="transcript-container">
            ${this.messages
							.map(
								(message) =>
									`<div class="message-container">
                    <span class="message-speaker">${message.speaker}:</span>
                    <span class="message-text">
						${formatCodeBlocks(message.text)}
					</span>
                  </div>`
							)
							.join("\n")}
              ${
								this.messageInProgress
									? `<div class="message-container message-in-progress">
                    <span class="message-speaker">${
											this.messageInProgress.speaker
										}:</span>
                    <span class="message-text">${formatCodeBlocks(
											this.messageInProgress.text
										)}█</span>
                </div>`
									: ""
							}
        </div>
        <div class="input-container">
            <form id="input-form">
                <textarea class="input" id="input" rows="1">${
									this.inputContents
								}</textarea>
            </form>
        </div>
    </div>
    <script>
        (function() {
            const inputEl = document.getElementById('input')
            inputEl.focus();
            inputEl.setSelectionRange(inputEl.value.length, inputEl.value.length);
            const vscode = acquireVsCodeApi();
            document.getElementById('input').addEventListener('keydown', function(e) {
                if (e.code === 'Enter' && !e.shiftKey) {
                    vscode.postMessage({
                        command: 'submit',
                        text: e.target.value,
                    });
                    document.getElementById('input').value = '';
                } else {
                    vscode.postMessage({
                        command: 'typed',
                        text: e.target.value,
                    });
                }
            }, false);
        })();
    </script>
    </body>
    </html>`;
	}
}
interface ChatCallbacks {
	onChange: (text: string) => void;
	onComplete: (text: string) => void;
	onError: (message: string, originalErrorEvent?: ErrorEvent) => void;
}
export class WSChatClient {
	static async new(addr: string): Promise<WSChatClient> {
		const wsclient = await WSClient.new<
			Omit<WSChatRequest, "requestId">,
			WSChatResponse
		>(addr);
		return new WSChatClient(wsclient);
	}

	constructor(
		private wsclient: WSClient<Omit<WSChatRequest, "requestId">, WSChatResponse>
	) {}

	chat(messages: Message[], callbacks: ChatCallbacks): void {
		this.wsclient.sendRequest(
			{
				kind: "request",
				messages,
			},
			(resp) => {
				switch (resp.kind) {
					case "response:change":
						callbacks.onChange(resp.message);
						return false;
					case "response:complete":
						callbacks.onComplete(resp.message);
						return true;
					case "response:error":
						callbacks.onError(resp.error);
						return false;
					default:
						return false;
				}
			}
		);
	}
}

function formatCodeBlocks(s: string) {
	const components = [];
	let remaining = s;
	let codeblockDelimiterCount = 0;
	while (remaining.length > 0) {
		const foundIndex = remaining.indexOf("```");
		if (foundIndex === -1) {
			components.push(remaining);
			if (codeblockDelimiterCount % 2 === 1) {
				components.push("</pre>");
			}
			break;
		}
		components.push(
			remaining.substring(0, foundIndex),
			codeblockDelimiterCount % 2 === 0 ? "<pre>" : "</pre>"
		);
		remaining = remaining.substring(foundIndex + 3) || "";
		codeblockDelimiterCount++;
	}
	const ret = components.join("");
	console.log(`original: ${s}`);
	console.log(`formatted: ${ret}`);
	return ret;
}
