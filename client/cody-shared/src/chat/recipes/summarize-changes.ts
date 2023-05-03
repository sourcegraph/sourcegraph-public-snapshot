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

        return null
    }
}
