import classNames from 'classnames'

import { MenuItem } from '@sourcegraph/wildcard'

import styles from './Recipes.module.css'

export interface RecipeActionProps {
    title: string
}

export const RecipeAction = ({ title }: RecipeActionProps): JSX.Element => {
    return (
        <MenuItem className={classNames(styles.recipeMenuWrapper)} onSelect={() => alert('Clicked!')}>
            {title}
        </MenuItem>
    )
}
