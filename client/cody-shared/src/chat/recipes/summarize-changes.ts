import { spawnSync } from 'child_process'
import * as path from 'path'

import * as vscode from 'vscode'

import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'
import { getDirectoryPathsBetween, stripPathPrefix } from './utils'

export class SummarizeChanges implements Recipe {
    public id = 'summarize-changes'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const rootDirPath = context.editor.getWorkspaceRootPath()
        if (!rootDirPath) {
            return null
        }

        // TODO: replace with context.editor
        const timeRangePicker = await vscode.window.createQuickPick()
        timeRangePicker.title = 'Over what time period would you like to summarize changes?'
        timeRangePicker.items = [{ label: 'last day' }, { label: 'last week' }, { label: 'since my last commit' }]
        timeRangePicker.onDidChangeValue(e => {
            console.log('# e', e)
        })
        await timeRangePicker.show()
        const timeRange = await new Promise(resolve => {
            timeRangePicker.onDidChangeSelection(items => {
                if (items.length > 0) {
                    resolve(items[0].label)
                } else {
                    resolve(null)
                }
            })
        })

        // Get scope
        const scopePicker = await vscode.window.createQuickPick()
        scopePicker.title = 'What set of code do you want to summarize changes over?'

        const filePath = context.editor.getActiveTextEditor()?.filePath
        let paths: string[] = []
        if (filePath) {
            for (const p of getDirectoryPathsBetween(rootDirPath, path.dirname(filePath))) {
                const pp = stripPathPrefix(rootDirPath, p)
                if (pp !== null) {
                    paths.push(pp)
                }
            }
        } else {
            paths.push(rootDirPath)
        }
        scopePicker.title = 'Over what scope would you like to summarize changes?'
        scopePicker.items = paths.map(p => ({ label: p }))
        await scopePicker.show()
        const scope = await new Promise(resolve => {
            scopePicker.onDidChangeSelection(items => {
                if (items.length > 0) {
                    resolve(items[0].label)
                } else {
                    resolve(null)
                }
            })
        })
        scopePicker.hide()

        // // invoke `git symbolic-ref refs/remotes/origin/HEAD` to get name of default branch
        // const defaultBranchName = spawnSync('git', ['symbolic-ref', 'refs/remotes/origin/HEAD'])

        // Human: Write the git command that prints the commits on branch "origin/main" in the time range "1 day ago" and covers changes to the directory "foo". Put the command in between <cmd></cmd> tags like this:

        // <cmd>
        // git ...
        // </cmd>

        // Assistant:  <cmd>
        // git log --since="1 day ago" -- foo origin/main

        console.log('#', timeRange, scope)
        let gitLLMOut = ''
        await new Promise<void>((resolve, reject) => {
            console.log(context.chatClient)
            context.chatClient.chat(
                [
                    {
                        speaker: 'human',
                        text: `Write the git command that prints the commits on branch "origin/main" in the time range "${timeRange}" and covers changes to the directory ${scope}. Put the command in between <cmd></cmd> tags like this:\n<cmd>\ngit ...\n</cmd>`,
                    },
                    {
                        speaker: 'assistant',
                        text: '',
                    },
                ],
                {
                    onChange: (text: string) => {
                        gitLLMOut = text
                    },
                    onComplete: () => {
                        resolve()
                    },
                    onError: (message: string, statusCode?: number) => {
                        reject(`error: ${message}, statusCode: ${statusCode}`)
                    },
                }
            )
        })
        gitLLMOut = gitLLMOut.trim()
        if (!gitLLMOut.endsWith('</cmd>') || !gitLLMOut.startsWith('<cmd>')) {
            vscode.window.showErrorMessage('bad command output:', gitLLMOut)
            return null
        }
        const gitCmd = gitLLMOut.substring('<cmd>'.length, -'</cmd>'.length).trim()

        // TODO: some validation of git command

        // execute gitCmd to get list of commits
        const gitCommitsOut = spawnSync('git', gitCmd.split(' '))

        // generate git command to get commits

        // group commits by author

        // for each author
        // for each commit for that author
        // for each file in the commit, summarize the change
        // recursively pop back up and summarize each level based on the summaries returned

        return null
    }
}
