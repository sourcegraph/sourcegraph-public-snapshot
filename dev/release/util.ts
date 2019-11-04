import * as readline from 'readline'
import { readFile, writeFile, mkdirp } from 'fs-extra'
import * as path from 'path'

export function addTime(dateTime: Date, numHours: number, numMinutes: number = 0): Date {
    const newDate = new Date(dateTime)
    newDate.setHours(dateTime.getHours() + numHours)
    newDate.setMinutes(dateTime.getMinutes() + numMinutes)
    return newDate
}

export async function readLine(prompt: string, cacheFile?: string): Promise<string> {
    if (!cacheFile) {
        return readLineNoCache(prompt)
    }

    try {
        return await readFile(cacheFile, { encoding: 'utf8' })
    } catch (err) {
        const userInput = await readLineNoCache(prompt)
        await mkdirp(path.dirname(cacheFile))
        await writeFile(cacheFile, userInput)
        return userInput
    }
}

function readLineNoCache(prompt: string): Promise<string> {
    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
    })
    return new Promise<string>(resolve =>
        rl.question(prompt, userInput => {
            rl.close()
            resolve(userInput)
        })
    )
}
