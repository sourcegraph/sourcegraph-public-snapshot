import { google, calendar_v3 } from 'googleapis'
import { OAuth2Client } from 'googleapis-common'
import open from 'open'
import { Credentials } from 'google-auth-library'
import { readLine } from './util'
import { readFile, writeFile } from 'mz/fs'

const SCOPES = ['https://www.googleapis.com/auth/calendar.events']
const TOKEN_PATH = '.secrets/google-calendar-token.json'

export async function getClient(): Promise<OAuth2Client> {
    const credentials = JSON.parse(
        await readLine(
            'Paste Google Calendar credentials (1Password "Release automation Google Calendar API App credentials"): ',
            '.secrets/google-calendar-credentials.json'
        )
    )
    const { client_secret, client_id, redirect_uris } = credentials.installed
    const oauth2Client = new OAuth2Client(client_id, client_secret, redirect_uris[0])
    oauth2Client.setCredentials(await getAccessToken(oauth2Client))
    return oauth2Client
}

async function getAccessToken(oauth2Client: OAuth2Client): Promise<Credentials> {
    try {
        const content = await readFile(TOKEN_PATH, { encoding: 'utf8' })
        // False positive https://github.com/typescript-eslint/typescript-eslint/issues/1269
        // eslint-disable-next-line @typescript-eslint/return-await
        return JSON.parse(content)
    } catch (err) {
        const token = await getAccessTokenNoCache(oauth2Client)
        await writeFile(TOKEN_PATH, JSON.stringify(token))
        return token
    }
}

async function getAccessTokenNoCache(oauth2Client: OAuth2Client): Promise<Credentials> {
    const authUrl = oauth2Client.generateAuthUrl({
        access_type: 'offline',
        scope: SCOPES,
    })
    await open(authUrl)
    const code = await readLine('Log in via the browser page that just opened and enter the code that appears: ')
    const token = await oauth2Client.getToken(code)
    return token.tokens
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
        },
    })
}

async function listEvents(auth: OAuth2Client): Promise<calendar_v3.Schema$Event[] | undefined> {
    const calendar = google.calendar({ version: 'v3', auth })
    const res = await calendar.events.list({
        calendarId: 'primary',
        timeMin: new Date().toISOString(),
        maxResults: 2500,
        singleEvents: true,
        orderBy: 'startTime',
    })
    return res.data.items
}
