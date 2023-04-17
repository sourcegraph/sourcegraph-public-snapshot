import classNames from 'classnames'

import { MenuItem } from '@sourcegraph/wildcard'

import styles from './Recipes.module.css'

export interface RecipeActionProps {
    title: string
    onClick?: () => void
}

export const RecipeAction = ({ title, onClick }: RecipeActionProps): JSX.Element => {
    return (
        <MenuItem
            className={classNames(styles.recipeMenuWrapper)}
            onSelect={() => {
                if (onClick) {
                    onClick()
                }
            }}
        >
            {title}
        </MenuItem>
    )
}
