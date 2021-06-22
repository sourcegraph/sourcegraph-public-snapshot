import * as path from 'path'
import * as readline from 'readline'
import { URL } from 'url'

import execa from 'execa'
import { readFile, writeFile, mkdir } from 'mz/fs'

/* eslint-disable @typescript-eslint/consistent-type-assertions */
export function formatDate(date: Date): string {
    return `${date.toLocaleString('en-US', {
        timeZone: 'UTC',
        dateStyle: 'medium',
        timeStyle: 'short',
    } as Intl.DateTimeFormatOptions)} (UTC)`
}
/* eslint-enable @typescript-eslint/consistent-type-assertions */

const addZero = (index: number): string => (index < 10 ? `0${index}` : `${index}`)

/**
 * Generates a link for comparing given Date with local time.
 */
export function timezoneLink(date: Date, linkName: string): string {
    const timeString = `${addZero(date.getUTCHours())}${addZero(date.getUTCMinutes())}`
    return `https://time.is/${timeString}_${date.getUTCDate()}_${date.toLocaleString('en-US', {
        month: 'short',
    })}_${date.getUTCFullYear()}_in_UTC?${encodeURI(linkName)}`
}

export const cacheFolder = './.secrets'

export async function readLine(prompt: string, cacheFile?: string): Promise<string> {
    if (!cacheFile) {
        return readLineNoCache(prompt)
    }

    try {
        return (await readFile(cacheFile, { encoding: 'utf8' })).trimEnd()
    } catch {
        const userInput = await readLineNoCache(prompt)
        await mkdir(path.dirname(cacheFile), { recursive: true })
        await writeFile(cacheFile, userInput)
        return userInput
    }
}

async function readLineNoCache(prompt: string): Promise<string> {
    const readlineInterface = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
    })
    const userInput = await new Promise<string>(resolve => readlineInterface.question(prompt, resolve))
    readlineInterface.close()
    return userInput
}

export function getWeekNumber(date: Date): number {
    const firstJan = new Date(date.getFullYear(), 0, 1)
    const day = 86400000
    return Math.ceil(((date.valueOf() - firstJan.valueOf()) / day + firstJan.getDay() + 1) / 7)
}

export function hubSpotFeedbackFormStub(version: string): string {
    const link = `[this feedback form](${hubSpotFeedbackFormURL(version)})`
    return `*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out ${link}.*`
}

function hubSpotFeedbackFormURL(version: string): string {
    const url = new URL('https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku')
    url.searchParams.set('update_version', version)

    return url.toString()
}

export async function ensureDocker(): Promise<execa.ExecaReturnValue<string>> {
    return execa('docker', ['version'], { stdout: 'ignore' })
}
