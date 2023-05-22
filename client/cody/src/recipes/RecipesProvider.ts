import * as vscode from 'vscode'

interface RecipeEntry {
    title: string
    icon?: string
}

const recipesList: { [key: string]: RecipeEntry } = {
    'explain-code-detailed': { title: 'Explain selected code (detailed)', icon: 'note' },
    'explain-code-high-level': { title: 'Explain selected code (high level)', icon: 'note' },
    'generate-unit-test': { title: 'Generate a unit test', icon: 'beaker' },
    'generate-docstring': { title: 'Generate a docstring', icon: 'book' },
    'improve-variable-names': { title: 'Improve variable names', icon: 'variable' },
    'translate-to-language': { title: 'Translate to different language' },
    'git-history': { title: 'Summarize recent code changes', icon: 'git-branch' },
    'find-code-smells': { title: 'Smell code', icon: 'warning' },
    fixup: { title: 'Fixup code from inline instructions', icon: 'zap' },
    'context-search': { title: 'Codebase context search', icon: 'search' },
    'release-notes': { title: 'Generate release notes', icon: 'preview' },
}

export class RecipesProvider implements vscode.TreeDataProvider<Dependency> {
    constructor() {}

    getTreeItem(element: Dependency): vscode.TreeItem {
        return element
    }

    getChildren(element?: Dependency): Thenable<Dependency[]> {
        return Promise.resolve(
            Object.entries(recipesList).map(
                ([id, data]) => new Dependency(data.title, id, new vscode.ThemeIcon(data.icon ?? 'zap'))
            )
        )
    }
}

class Dependency extends vscode.TreeItem {
    constructor(public readonly label: string, public description: string, public iconPath: vscode.ThemeIcon) {
        super(label, vscode.TreeItemCollapsibleState.None)
        this.command = {
            title: 'Run Recipe From Tree View',
            command: 'cody.recipes.run-from-tree-view',
            arguments: [description],
        }
    }
}
