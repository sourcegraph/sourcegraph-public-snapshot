import { createServer } from "http";
import { parse } from 'url';
import { WebSocketServer } from 'ws';
import express from "express";
import * as bodyParser from 'body-parser';
import { OpenAIBackend, langKeywordStopStrings, prompt_refs_history_inlinecomments } from "./prompts/openai";
import * as openai from "openai";
import { WSChatRequest, WSChatResponseChange, WSChatResponseComplete, WSChatResponseError, WSCompletionResponseCompletion, WSCompletionsRequest, WSCompletionResponse } from "@sourcegraph/cody-common";
import { ClaudeBackend } from "./prompts/claude";
import { defaultModelParams } from "@completion/sampling";

const openaiKey = process.env.OPENAI_KEY;
if (!openaiKey) {
	throw new Error("OPENAI_KEY missing");
}
const claudeKey = process.env.CLAUDE_KEY;
if (!claudeKey) {
	throw new Error("CLAUDE_KEY missing");
}
const port = process.env.CODY_PORT || '8080';

const openaiConfig = new openai.Configuration({
	apiKey: openaiKey,
});
const openaiBackend = new OpenAIBackend(
	openaiConfig,
	{
		// eslint-disable-next-line @typescript-eslint/naming-convention
		max_tokens: 256,
		model: "code-cushman-001",
		n: 2,
	},
	prompt_refs_history_inlinecomments(1000),
	langKeywordStopStrings
);
const claudeBackend = new ClaudeBackend(
	claudeKey,
	{
		...defaultModelParams,
		name: "santa-h-v3-s400",
		temperature: 0.2,
	},
)

const app = express();
app.use(bodyParser.json());

const httpServer = createServer(app);
const wssCompletions = new WebSocketServer({ noServer: true });
wssCompletions.on('connection', ws => {
  console.log("completions:connection");
  ws.on('message', async data => {
	console.log("completions:request")
	const req = JSON.parse(data.toString()) as WSCompletionsRequest
	if (req.kind !== 'getCompletions' || !req.args || !req.requestId) {
		console.error(`invalid request ${data.toString()}`)
		return;
	}
	const {completions, debug} = await openaiBackend.getCompletions(req.args)
	const response: WSCompletionResponseCompletion = {
		completions,
		requestId: req.requestId,
		kind: 'completion',
		debugInfo: debug,
	}
	ws.send(JSON.stringify(response), err => {
		if (err) {
			console.error(`failed to send websocket getCompletions response: ${err}`)
			return
		}
		const doneMsg: WSCompletionResponse = {requestId: req.requestId, kind: 'done'}
		ws.send(JSON.stringify(doneMsg))
	})
  });
// TODO(beyang): handle shutdown
});

const wssChat = new WebSocketServer({ noServer: true });
wssChat.on('connection', ws => {
	console.log("chat:connection")
	// TODO(beyang): Close connection after timeout. Probably should keep connection around,
	// rather than closing after every response?

	ws.on("message", async (data) => {
		console.log("chat:request")
		const req = JSON.parse(data.toString()) as WSChatRequest;
		if (!req.requestId || !req.messages) {
			console.error(`invalid request ${data.toString()}`);
			return;
		}
		claudeBackend.chat(req.messages, {
			onChange: (message) => {
				const msg: WSChatResponseChange = {
					requestId: req.requestId,
					kind: "response:change",
					message,
				};
				ws.send(JSON.stringify(msg));
			},
			onComplete: (message) => {
				const msg: WSChatResponseComplete = {
					requestId: req.requestId,
					kind: "response:complete",
					message,
				};
				ws.send(JSON.stringify(msg), (err) => {
					if (err) {
						console.error(`error sending last response message: ${err}`);
					}
				});
			},
			onError: (error) => {
				const msg: WSChatResponseError = {
					requestId: req.requestId,
					kind: "response:error",
					error,
				};
				ws.send(JSON.stringify(msg), (err) => {
					if (err) {
						console.error(`error sending error message: ${err}`);
					}
				});
			},
		});
	});
})

httpServer.on('upgrade', (request, socket, head) => {
	if (!request.url) {
		return
	}
	const { pathname } = parse(request.url);
	if (pathname === '/completions') {
	  wssCompletions.handleUpgrade(request, socket, head, function done(ws) {
		wssCompletions.emit('connection', ws, request);
	  });
	} else if (pathname === '/chat') {
	  wssChat.handleUpgrade(request, socket, head, function done(ws) {
		wssChat.emit('connection', ws, request);
	  });
	} else {
	  socket.destroy();
	}
})

console.log(`Server listening on :${port}`);
httpServer.listen(port);
