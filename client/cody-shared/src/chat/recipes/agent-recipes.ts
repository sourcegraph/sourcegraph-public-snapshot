import { ChatQuestion } from './chat-question'
// import { ContextSearch } from './context-search'
import { ExplainCodeDetailed } from './explain-code-detailed'
import { ExplainCodeHighLevel } from './explain-code-high-level'
// import { FileFlow } from './file-flow'
import { FindCodeSmells } from './find-code-smells'
// import { Fixup } from './fixup'
import { GenerateDocstring } from './generate-docstring'
// import { ReleaseNotes } from './generate-release-notes'
import { GenerateTest } from './generate-test'
// import { GitHistory } from './git-log'
import { ImproveVariableNames } from './improve-variable-names'
// import { InlineAssist } from './inline-chat'
// import { NextQuestions } from './next-questions'
// import { NonStop } from './non-stop'
// import { OptimizeCode } from './optimize-code'
import { Recipe, RecipeID } from './recipe'
import { TranslateToLanguage } from './translate'

function nullLog(filterLabel: string, text: string, ...args: unknown[]): void {
    // Do nothing
}

export const registeredRecipes: { [id in RecipeID]?: Recipe } = {}

export function getRecipe(id: RecipeID): Recipe | undefined {
    return registeredRecipes[id]
}

function registerRecipe(id: RecipeID, recipe: Recipe): void {
    registeredRecipes[id] = recipe
}

function init(): void {
    if (Object.keys(registeredRecipes).length > 0) {
        return
    }

    const recipes: Recipe[] = [
        new ChatQuestion(nullLog),
        new ExplainCodeDetailed(),
        new ExplainCodeHighLevel(),
        new GenerateDocstring(),
        new GenerateTest(),
        new ImproveVariableNames(),
        new TranslateToLanguage(),
        new FindCodeSmells(),
    ]

    for (const recipe of recipes) {
        const existingRecipe = getRecipe(recipe.id)
        if (existingRecipe) {
            throw new Error(`Duplicate recipe with ID ${recipe.id}`)
        }
        registerRecipe(recipe.id, recipe)
    }
}

init()
