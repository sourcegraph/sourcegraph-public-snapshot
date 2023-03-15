import { execFile } from 'child_process'
import * as path from 'path'

import { Editor } from '../../editor'
import { truncateText } from '../prompt'

import { Recipe, RecipePrompt } from './recipe'

export class GitHistory implements Recipe {
    public getID(): string {
        return 'gitHistory'
    }

    public async getPrompt(maxTokens: number, editor: Editor): Promise<RecipePrompt | null> {
        const activeEditor = editor.getActiveTextEditor()
        if (!activeEditor) {
            return null
        }
        const dirPath = path.dirname(activeEditor.filePath)

        const items = [
            {
                label: 'Last 5 items',
                args: ['log', '-n5', '--stat'],
                displayText: 'What changed in my codebase in the last 5 commits?',
            },
            {
                label: 'Last day',
                args: ['log', '--since', '1 day', '--stat'],
                displayText: 'What has changed in my codebase in the last day?',
            },
            {
                label: 'Last week',
                args: ['log', "--since='1 week'", "--pretty='%an: %s'"],
                displayText: 'What changed in my codebase in the last week?',
            },
        ]
        const selectedLabel = await editor.showQuickPick(items.map(e => e.label))
        if (!selectedLabel) {
            return null
        }
        const selected = Object.fromEntries(
            items.map(({ label, args, displayText }) => [label, { args, displayText }])
        )[selectedLabel]
        const { args: gitArgs, displayText } = selected

        const gitLogRaw = await new Promise<string>((resolve, reject) => {
            execFile('git', ['--no-pager', ...gitArgs], { cwd: dirPath }, (error, stdout, stderr) => {
                if (error) {
                    reject(`git ${gitArgs.join(' ')} error: ${error}\nstderr:${stderr}`)
                    return
                }
                resolve(stdout)
            })
        })

        const gitlogStr = gitLogRaw.toLocaleString()
        const messageTextTempl =
            'Summarize these commits:\n{gitlogStr}\n\nProvide your response in the form of a bulleted list. Do not mention the commit hashes.'
        let messageText = messageTextTempl.replace('{gitlogStr}', gitlogStr)
        if (messageText.length > maxTokens * 3.25) {
            const truncatedGitLog = truncateText(
                gitlogStr,
                Math.floor(maxTokens * 3.25) - messageTextTempl.length + '{gitlogStr}'.length
            )
            messageText = messageTextTempl.replace('{gitlogStr}', truncatedGitLog)
            void editor.showWarningMessage('Truncated extra long git log output, so summary may be incomplete')
        }

        return {
            displayText,
            promptMessage: {
                speaker: 'you',
                text: messageText,
            },
            contextMessages: [],
            botResponsePrefix: 'Here is a summary of recent changes:\n- ',
        }
    }
}
