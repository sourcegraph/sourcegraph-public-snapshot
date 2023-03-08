import { ExplainCodeDetailed, ExplainCodeHighLevel } from './explainCode'
import { GenerateDocstring } from './generateDocstring'
import { GenerateTest } from './generateTest'
import { GitHistory } from './gitLog'
import { ImproveVariableNames } from './improveVariableNames'
import { Recipe } from './recipe'
import { TranslateToLanguage } from './translate'

const registeredRecipes: { [id: string]: Recipe } = {}

export function registerRecipe(id: string, recipe: Recipe) {
	registeredRecipes[id] = recipe
}

export function getRecipe(id: string): Recipe | null {
	return registeredRecipes[id]
}

function init() {
	const recipes: Recipe[] = [
		new ExplainCodeDetailed(),
		new ExplainCodeHighLevel(),
		new GenerateDocstring(),
		new GenerateTest(),
		new GitHistory(),
		new ImproveVariableNames(),
		new TranslateToLanguage(),
	]

	for (const recipe of recipes) {
		const existingRecipe = getRecipe(recipe.getID())
		if (existingRecipe) {
			throw new Error(`Duplicate recipe with ID ${recipe.getID()}`)
		}
		registerRecipe(recipe.getID(), recipe)
	}
}

init()
