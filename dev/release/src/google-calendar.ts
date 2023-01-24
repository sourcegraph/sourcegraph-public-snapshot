import { createServer } from 'http'
import type { IncomingMessage, Server, ServerResponse } from 'http'
import { AddressInfo } from 'net'

import { addMinutes } from 'date-fns'
import { Credentials } from 'google-auth-library'
import { google, calendar_v3 } from 'googleapis'
import { OAuth2Client } from 'googleapis-common'
import { readFile, writeFile } from 'mz/fs'
import open from 'open'

import { readLine, cacheFolder } from './util'

const SCOPES = ['https://www.googleapis.com/auth/calendar.events']
const TOKEN_PATH = `${cacheFolder}/google-calendar-token.json`

export async function getClient(): Promise<OAuth2Client> {
    const credentials = JSON.parse(
        await readLine(
            'Paste Google Calendar credentials (1Password "Release automation Google Calendar API App credentials"): ',
            `${cacheFolder}/google-calendar-credentials.json`
        )
    )
    const { client_secret, client_id } = credentials.installed
    const server = await new Promise<Server>(resolve => {
        const s = createServer()
        s.listen(0, () => resolve(s))
    })
    const { port } = server.address() as AddressInfo
    const oauth2Client = new OAuth2Client({
        clientId: client_id,
        clientSecret: client_secret,
        redirectUri: `http://localhost:${port}`,
    })
    oauth2Client.setCredentials(await getAccessToken(server, oauth2Client))
    server.close()
    return oauth2Client
}

async function getAccessToken(server: Server, oauth2Client: OAuth2Client): Promise<Credentials> {
    try {
        const content = await readFile(TOKEN_PATH, { encoding: 'utf8' })
        return JSON.parse(content)
    } catch {
        const token = await getAccessTokenNoCache(server, oauth2Client)
        await writeFile(TOKEN_PATH, JSON.stringify(token))
        return token
    }
}

async function getAccessTokenNoCache(server: Server, oauth2Client: OAuth2Client): Promise<Credentials> {
    const authUrl = oauth2Client.generateAuthUrl({
        access_type: 'offline',
        scope: SCOPES,
    })

    const authCode = await new Promise<string>((resolve, reject) => {
        server.on('request', (request: IncomingMessage, response: ServerResponse) => {
            const urlParts = new URL(request.url ?? '', 'http://localhost').searchParams
            const code = urlParts.get('code')
            const error = urlParts.get('error')
            if (code) {
                resolve(code)
            } else {
                reject(error)
            }

            response.end('Authentication successful! Please return to the console')
        })
        ;(async () => open(authUrl, { wait: false }).then(cp => cp.unref()))()
    })

    const { tokens } = await oauth2Client.getToken(authCode)
    return tokens
}

export interface EventOptions {
    anyoneCanAddSelf?: boolean
    attendees?: string[]
    startDate?: string
    endDate?: string
    startDateTime?: string
    endDateTime?: string
    description?: string
    title: string
    transparency: string
}

export async function ensureEvent(
    {
        anyoneCanAddSelf = false,
        attendees = [],
        startDate,
        endDate,
        startDateTime,
        endDateTime,
        description = '',
        title,
        transparency,
    }: EventOptions,
    auth: OAuth2Client
): Promise<void> {
    const existingEvents = await listEvents(auth)
    const foundEvents = (existingEvents || []).filter(({ summary }) => summary === title)
    if (foundEvents.length > 0) {
        console.log(`Event ${JSON.stringify(title)} already exists (not updating)`)
        return
    }

    const calendar = google.calendar({ version: 'v3', auth })
    await calendar.events.insert({
        calendarId: 'primary',
        requestBody: {
            anyoneCanAddSelf,
            attendees: attendees.map(email => ({ email })),
            start: { date: startDate, dateTime: startDateTime },
            end: { date: endDate, dateTime: endDateTime },
            description,
            summary: title,
            transparency,
        },
    })
}

async function listEvents(auth: OAuth2Client): Promise<calendar_v3.Schema$Event[] | undefined> {
    const calendar = google.calendar({ version: 'v3', auth })
    const result = await calendar.events.list({
        calendarId: 'primary',
        timeMin: new Date().toISOString(),
        maxResults: 2500,
        singleEvents: true,
        orderBy: 'startTime',
    })
    return result.data.items
}

export function calendarTime(date: string): { startDateTime: string; endDateTime: string } {
    return {
        startDateTime: new Date(date).toISOString(),
        endDateTime: addMinutes(new Date(date), 1).toISOString(),
    }
}
