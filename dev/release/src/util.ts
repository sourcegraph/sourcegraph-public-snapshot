import * as readline from 'readline'
import { readFile, writeFile } from 'mz/fs'
import mkdirp from 'mkdirp-promise'
import * as path from 'path'

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

async function readLineNoCache(prompt: string): Promise<string> {
    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
    })
    const userInput = await new Promise<string>(resolve => rl.question(prompt, resolve))
    rl.close()
    return userInput
}
