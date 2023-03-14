import { createServer } from 'http'
import { parse } from 'url'

import * as bodyParser from 'body-parser'
import express from 'express'
import { GoogleSpreadsheet } from 'google-spreadsheet'
import { WebSocketServer } from 'ws'

import {
    Feedback,
    Message,
    WSChatRequest,
    WSChatResponseChange,
    WSChatResponseComplete,
    WSChatResponseError,
    feedbackToSheetRow,
} from '@sourcegraph/cody-common'

import { authenticate, getUsers, User } from './auth'
import { wsHandleGetCompletions } from './completions'
import { getInfo } from './info'
import { ClaudeBackend } from './prompts/claude'

const anthropicApiKey = process.env.ANTHROPIC_API_KEY
if (!anthropicApiKey) {
    throw new Error('ANTHROPIC_API_KEY is missing')
}

const usersPath = process.env.CODY_USERS_PATH
if (!usersPath) {
    throw new Error('CODY_USERS_PATH is missing')
}

const port = process.env.CODY_PORT || '8080'

const feedbackServiceAccount = process.env.FEEDBACK_SERVICE_ACCOUNT
const feedbackServiceAccountKey: string = (process.env.FEEDBACK_SERVICE_ACCOUNT_KEY || '').replace(/\\n/gm, '\n')
const feedbackSheetID = process.env.FEEDBACK_SHEET_ID
const feedbackSheetTitle = process.env.FEEDBACK_SHEET_TITLE
const telemetrySheetTitle = process.env.TELEMETRY_SHEET_TITLE
const telemetryVersion = 'v0'

if (!feedbackSheetID || !feedbackServiceAccount || !feedbackServiceAccountKey || !feedbackSheetTitle) {
    console.error('feedback disabled')
}
if (!feedbackSheetID || !feedbackServiceAccount || !feedbackServiceAccountKey || !telemetrySheetTitle) {
    console.error('telemetry disabled')
}

// Character length of this preamble is 806 chars or ~230 tokens (at a conservative rate of 3.5 chars per token).
// If this is modified, then `PROMPT_PREAMBLE_LENGTH` in prompt.ts should be updated.
const codyPreambleMessages: Message[] = [
    {
        speaker: 'you',
        text: `You are Cody, an AI-powered coding assistant created by Sourcegraph that performs the following actions:
- Answer general programming questions
- Answer questions about code that I have provided to you
- Generate code that matches a written description
- Explain what a section of code does

In your responses, you should obey the following rules:
- Be as brief and concise as possible without losing clarity
- Any code snippets should be markdown-formatted (placed in-between triple backticks like this "\`\`\`").
- Answer questions only if you know the answer or can make a well-informed guess. Otherwise, tell me you don't know, and tell me what context I need to provide to you in order for you to answer the question.
- Do not reference any file names or URLs, unless you are sure they exist.`,
    },
    {
        speaker: 'bot',
        text: 'Understood. I am Cody, an AI-powered coding assistant created by Sourcegraph and will follow the rules above',
    },
]

const claudeBackend = new ClaudeBackend(
    anthropicApiKey,
    {
        model: 'claude-v1',
        temperature: 0.2,
        stop_sequences: ['\n\nHuman:'],
        max_tokens_to_sample: 1000,
        top_p: -1,
        top_k: -1,
    },
    codyPreambleMessages
)

const shortAnswerBackend = new ClaudeBackend(
    anthropicApiKey,
    {
        model: 'claude-v1',
        temperature: 0,
        stop_sequences: ['\n\nHuman:'],
        max_tokens_to_sample: 1,
        top_p: -1,
        top_k: -1,
    },
    []
)

const app = express()
app.use(bodyParser.json({ limit: '50mb' }))

const httpServer = createServer(app)

const wssCompletions = new WebSocketServer({ noServer: true })
wssCompletions.on('connection', ws => {
    console.log('completions:connection')
    ws.on('message', async data => {
        try {
            console.log('completions:request')
            const req = JSON.parse(data.toString())
            switch (req.kind) {
                case 'getCompletions':
                    if (req.kind !== 'getCompletions' || !req.args || !req.requestId) {
                        console.error(`invalid request ${data.toString()}`)
                        return
                    }
                    await wsHandleGetCompletions(ws, req)
                    return
                default:
                    console.error(`invalid request ${data.toString()}`)
                    return
            }
        } catch (error: any) {
            console.error('Uncaught error', error)
        }
    })
    // TODO(beyang): handle shutdown
})

app.get('/info', (req, res) => {
    const user = authenticate(req.headers.authorization, null, getUsers(usersPath))
    if (!user) {
        res.status(401).send({ error: 'unauthorized' })
        return
    }

    const { q } = req.query
    getInfo(shortAnswerBackend, q as string)
        .then(info => {
            res.send(JSON.stringify(info))
        })
        .catch(error => {
            res.status(500).send({ error })
        })
})

app.post('/feedback', (req, res) => {
    const user = authenticate(req.headers.authorization, null, getUsers(usersPath))
    if (!user) {
        res.status(401).send({ error: 'unauthorized' })
        return
    }

    const feedback = req.body as Feedback
    try {
        postFeedback(user, feedback).finally(() => {
            res.send({ success: true })
        })
    } catch (error) {
        res.status(500).send({ success: false, error })
    }
})

async function postAction(user: User | null, data: any): Promise<void> {
    if (!feedbackSheetID || !feedbackServiceAccount || !feedbackServiceAccountKey || !telemetrySheetTitle) {
        return
    }
    const event = {
        name: user?.name,
        email: user?.email,
        ...data,
    }

    try {
        const doc = new GoogleSpreadsheet(feedbackSheetID)
        await doc.useServiceAccountAuth({
            client_email: feedbackServiceAccount,
            private_key: feedbackServiceAccountKey,
        })

        await doc.loadInfo()
        const sheet = doc.sheetsByTitle[telemetrySheetTitle]
        if (!sheet) {
            throw new Error(`sheet title "${telemetrySheetTitle}" not found`)
        }
        await sheet.addRow({
            name: user?.name || '',
            email: user?.email || '',
            event: JSON.stringify(data),
            timestamp: new Date().toISOString(),
            telemetryVersion,
        })
    } catch (error) {
        console.error('postAction error', error)
    }
}

async function postFeedback(user: User, feedback: Feedback): Promise<void> {
    if (!feedbackSheetID || !feedbackServiceAccount || !feedbackServiceAccountKey || !feedbackSheetTitle) {
        return
    }
    feedback.user = user.email

    try {
        const doc = new GoogleSpreadsheet(feedbackSheetID)
        await doc.useServiceAccountAuth({
            client_email: feedbackServiceAccount,
            private_key: feedbackServiceAccountKey,
        })

        await doc.loadInfo()
        const sheet = doc.sheetsByTitle[feedbackSheetTitle]
        if (!sheet) {
            throw new Error(`sheet title "${feedbackSheetTitle}" not found`)
        }
        await sheet.addRow(feedbackToSheetRow(feedback))
    } catch (error) {
        console.error('postFeedback error', error)
    }
}

const wssChat = new WebSocketServer({ noServer: true })
wssChat.on('connection', (ws, initReq) => {
    if (!initReq.url) {
        console.error('error: expected request url to be non-empty')
        return
    }
    const { search } = parse(initReq.url)
    const user = authenticate(
        initReq.headers.authorization,
        new URLSearchParams(search || '').get('access_token'),
        getUsers(usersPath)
    )

    console.log('chat:connection')
    // TODO(beyang): Close connection after timeout. Probably should keep connection around,
    // rather than closing after every response?

    ws.on('message', async data => {
        console.log('chat:request')
        const req = JSON.parse(data.toString()) as WSChatRequest
        if (!req.requestId || !req.messages) {
            console.error(`invalid request ${data.toString()}`)
            return
        }

        postAction(user, { event: req.metadata }) // telemetry

        claudeBackend.chat(req.messages, {
            onChange: message => {
                const msg: WSChatResponseChange = { requestId: req.requestId, kind: 'response:change', message }
                ws.send(JSON.stringify(msg))
            },
            onComplete: message => {
                const msg: WSChatResponseComplete = { requestId: req.requestId, kind: 'response:complete', message }
                ws.send(JSON.stringify(msg), err => {
                    if (err) {
                        console.error(`error sending last response message: ${err}`)
                    }
                })
            },
            onError: error => {
                const msg: WSChatResponseError = { requestId: req.requestId, kind: 'response:error', error }
                ws.send(JSON.stringify(msg), err => {
                    if (err) {
                        console.error(`error sending error message: ${err}`)
                    }
                })
            },
        })
    })
})

httpServer.on('upgrade', (request, socket, head) => {
    if (!request.url) {
        return
    }

    const { pathname, search } = parse(request.url)

    const user = authenticate(
        request.headers.authorization,
        new URLSearchParams(search || '').get('access_token'),
        getUsers(usersPath)
    )
    if (!user) {
        socket.end('HTTP/1.1 401 Unauthorized\r\n\r\n')
        return
    }

    if (pathname === '/completions') {
        wssCompletions.handleUpgrade(request, socket, head, ws => {
            wssCompletions.emit('connection', ws, request)
        })
    } else if (pathname === '/chat') {
        wssChat.handleUpgrade(request, socket, head, ws => {
            wssChat.emit('connection', ws, request)
        })
    } else {
        socket.destroy()
    }
})

console.log(`Server listening on :${port}`)
httpServer.listen(port)
