import { createServer, type IncomingMessage, type Server, type ServerResponse } from 'http'
import type { AddressInfo } from 'net'

import { addMinutes } from 'date-fns'
import type { Credentials } from 'google-auth-library'
import { google, type calendar_v3 } from 'googleapis'
import { OAuth2Client } from 'googleapis-common'
import { DateTime } from 'luxon'
import { readFile, writeFile } from 'mz/fs'
import open from 'open'

import { readLine, cacheFolder } from './util'

export interface Installed {
    client_id?: string
    client_secret?: string
    redirect_uri?: string
}
export interface OAuth2ClientOptions {
    installed: Installed
}

const SCOPES = ['https://www.googleapis.com/auth/calendar.events']
const TOKEN_PATH = `${cacheFolder}/google-calendar-token.json`

export async function getClient(): Promise<OAuth2Client> {
    const credentials: OAuth2ClientOptions = JSON.parse(
        await readLine(
            'Paste Google Calendar credentials (1Password "Release automation Google Calendar API App credentials"): ',
            `${cacheFolder}/google-calendar-credentials.json`
        )
    )
    const oauth2Client = await authorize(credentials)
    return oauth2Client
}
async function authorize(credentials: OAuth2ClientOptions): Promise<OAuth2Client> {
    let oauth2Client: OAuth2Client
    try {
        const token = await getAccessCachedToken()
        oauth2Client = new OAuth2Client({
            clientId: credentials.installed.client_id,
            clientSecret: credentials.installed.client_secret,
            redirectUri: credentials.installed.redirect_uri,
        })
        oauth2Client.setCredentials(token)
        return oauth2Client
    } catch {
        const server = await new Promise<Server>(resolve => {
            const serv = createServer()
            serv.listen(0, () => resolve(serv))
        })
        const { port } = server.address() as AddressInfo
        const oauth2Client = new OAuth2Client({
            clientId: credentials.installed.client_id,
            clientSecret: credentials.installed.client_secret,
            redirectUri: `http://localhost:${port}`,
        })

        const token = await getAccessTokenNoCache(server, oauth2Client)
        await writeFile(TOKEN_PATH, JSON.stringify(token))
        oauth2Client.setCredentials(token)
        server.close()
        return oauth2Client
    }
}

async function getAccessCachedToken(): Promise<Credentials> {
    const content = await readFile(TOKEN_PATH, { encoding: 'utf8' })
    return JSON.parse(content)
}

async function getAccessTokenNoCache(server: Server, oauth2Client: OAuth2Client): Promise<Credentials> {
    const authUrl = oauth2Client.generateAuthUrl({
        access_type: 'offline',
        scope: SCOPES,
    })

    const authCode = await new Promise<string>((resolve, reject) => {
        server.on('request', (request: IncomingMessage, response: ServerResponse) => {
            try {
                const urlParts = new URL(request.url ?? '', 'http://localhost').searchParams
                const code = urlParts.get('code')
                const error = urlParts.get('error')
                if (error) {
                    throw new Error(error)
                }
                if (code) {
                    resolve(code)
                }
                response.end('Authentication successful! Please return to the console')
            } catch (error) {
                reject(error)
            }
        })
        open(authUrl, { wait: false })
            .then(childProcess => childProcess.unref())
            .catch(reject)
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
            attendees: attendees.map(email => ({ email, optional: true })),
            start: { date: startDate, dateTime: startDateTime },
            end: { date: endDate, dateTime: endDateTime },
            description,
            summary: title,
            transparency,
        },
    })
}

export async function listEvents(auth: OAuth2Client): Promise<calendar_v3.Schema$Event[] | undefined> {
    const calendar = google.calendar({ version: 'v3', auth })
    const result = await calendar.events.list({
        calendarId: 'primary',
        timeMin: new Date().toISOString(),
        timeMax: DateTime.now().plus({ year: 1 }).toJSDate().toISOString(), // this ends up returning a lot of events, so filtering down to the next year should be fine
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
