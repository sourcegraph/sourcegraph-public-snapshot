import { ChatQuestion } from './chat-question'
import { ContextSearch } from './context-search'
import { ExplainCodeDetailed } from './explain-code-detailed'
import { ExplainCodeHighLevel } from './explain-code-high-level'
import { FindCodeSmells } from './find-code-smells'
import { Fixup } from './fixup'
import { GenerateDocstring } from './generate-docstring'
import { ReleaseNotes } from './generate-release-notes'
import { GenerateTest } from './generate-test'
import { GitHistory } from './git-log'
import { ImproveVariableNames } from './improve-variable-names'
import { InlineChat } from './inline-chat'
import { NextQuestions } from './next-questions'
import { Recipe, RecipeID } from './recipe'
import { TranslateToLanguage } from './translate'

const registeredRecipes: { [id in RecipeID]?: Recipe } = {}

export function registerRecipe(id: RecipeID, recipe: Recipe): void {
    registeredRecipes[id] = recipe
}

export function getRecipe(id: RecipeID): Recipe | undefined {
    return registeredRecipes[id]
}

function init(): void {
    if (Object.keys(registeredRecipes).length > 0) {
        return
    }

    const recipes: Recipe[] = [
        new ChatQuestion(),
        new ExplainCodeDetailed(),
        new ExplainCodeHighLevel(),
        new InlineChat(),
        new GenerateDocstring(),
        new GenerateTest(),
        new GitHistory(),
        new ImproveVariableNames(),
        new Fixup(),
        new TranslateToLanguage(),
        new FindCodeSmells(),
        new NextQuestions(),
        new ContextSearch(),
        new ReleaseNotes(),
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
