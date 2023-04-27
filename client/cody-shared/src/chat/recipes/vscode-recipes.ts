import { ChatQuestion } from './chat-question'
import { ExplainCodeDetailed } from './explain-code-detailed'
import { ExplainCodeHighLevel } from './explain-code-high-level'
import { FindCodeSmells } from './find-code-smells'
import { Fixup } from './fixup'
import { GenerateDocstring } from './generate-docstring'
import { GenerateTest } from './generate-test'
import { GitHistory } from './git-log'
import { ImproveVariableNames } from './improve-variable-names'
import { NextQuestions } from './next-questions'
import { Recipe } from './recipe'
import { TranslateToLanguage } from './translate'

const registeredRecipes: { [id: string]: Recipe } = {}

export function registerRecipe(id: string, recipe: Recipe): void {
    registeredRecipes[id] = recipe
}

export function getRecipe(id: string): Recipe | null {
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
        new GenerateDocstring(),
        new GenerateTest(),
        new GitHistory(),
        new ImproveVariableNames(),
        new Fixup(),
        new TranslateToLanguage(),
        new FindCodeSmells(),
        new NextQuestions(),
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
