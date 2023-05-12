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
import { Recipe } from './recipe'
import { TranslateToLanguage } from './translate'

export interface ChatOptions {
    numCodeResults?: number
    numTextResults?: number
}

export interface Options {
    chat?: ChatOptions
}

export class VSCodeRecipeRegistry {
    private registeredRecipes: Record<string, Recipe> = {}
    constructor(options: Options) {
        const recipes: Recipe[] = [
            new ChatQuestion(options.chat?.numCodeResults, options.chat?.numTextResults),
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
            const existingRecipe = this.getRecipe(recipe.id)
            if (existingRecipe) {
                throw new Error(`Duplicate recipe with ID ${recipe.id}`)
            }
            this.registeredRecipes[recipe.id] = recipe
        }
    }

    public getRecipe(id: string): Recipe | null {
        return this.registeredRecipes[id]
    }
}
