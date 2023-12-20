import { FC, ReactNode } from 'react'

import classNames from 'classnames'

import { Filter } from '@sourcegraph/shared/src/search/stream'
import { Button, Code } from '@sourcegraph/wildcard'

import styles from './RecipesLists.module.scss'

enum CodeRecipesType {
    RemoveGenFiles,
    RemoveTestFiles,
    CountAll,
}

interface CodeFilterRecipesProps {
    values: CodeRecipesType[]
    onChange: (nextValues: CodeRecipesType[]) => void
}

export const CodeFilterRecipes: FC<CodeFilterRecipesProps> = props => {
    const { values, onChange } = props

    const handleRecipeClick = (value: CodeRecipesType): void => {
        const recipeIndex = values.findIndex(recipe => recipe === value)

        if (recipeIndex !== -1) {
            // Remove already included recipe (turn it off) from the filter recipes
            onChange(values.filter(recipe => recipe !== value))
        } else {
            // Include (turn on) some recipe to the active filter recipes
            onChange([...values, value])
        }
    }

    return (
        <ul className={styles.recipesList}>
            <li>
                <RecipeButton
                    selected={values.includes(CodeRecipesType.RemoveGenFiles)}
                    onClick={() => handleRecipeClick(CodeRecipesType.RemoveGenFiles)}
                >
                    Remove generated files
                    <Code>-file:^gen -file:^\.gen</Code>
                </RecipeButton>
            </li>

            <li>
                <RecipeButton
                    selected={values.includes(CodeRecipesType.RemoveTestFiles)}
                    onClick={() => handleRecipeClick(CodeRecipesType.RemoveTestFiles)}
                >
                    Remove test files
                    <Code>-file:test</Code>
                </RecipeButton>
            </li>

            <li>
                <RecipeButton
                    selected={values.includes(CodeRecipesType.CountAll)}
                    onClick={() => handleRecipeClick(CodeRecipesType.CountAll)}
                >
                    Show all instances
                    <Code>count:all</Code>
                </RecipeButton>
            </li>
        </ul>
    )
}

interface UtilitiesFilterRecipesProps {
    filters: Filter[]
    filtersQuery: string
    onFilterChange: (nextQuery: string) => void
}

export const UtilitiesFilterRecipes: FC<UtilitiesFilterRecipesProps> = props => {
    const { filtersQuery, filters, onFilterChange } = props

    const handleRecipeClick = (filter: Filter): void => {
        const alreadyIncluded = filtersQuery.includes(filter.value)

        if (alreadyIncluded) {
            onFilterChange(filtersQuery.replaceAll(filter.value, ''))
        } else {
            onFilterChange(`${filtersQuery} ${filter.value}`)
        }
    }

    if (filters.length === 0) {
        return null
    }

    return (
        <ul className={styles.recipesList}>
            {filters.map(filter => (
                <RecipeButton
                    key={filter.value}
                    selected={filtersQuery.includes(filter.value)}
                    onClick={() => handleRecipeClick(filter)}
                >
                    {filter.label}
                </RecipeButton>
            ))}
        </ul>
    )
}

interface RecipeButtonProps {
    selected: boolean
    children: ReactNode
    onClick?: () => void
}

const RecipeButton: FC<RecipeButtonProps> = props => {
    const { selected, children, onClick } = props

    return (
        <Button
            variant={selected ? 'primary' : 'secondary'}
            outline={!selected}
            className={classNames(styles.recipesListItem, { [styles.recipesListItemSelected]: selected })}
            onClick={onClick}
        >
            {children}
        </Button>
    )
}
