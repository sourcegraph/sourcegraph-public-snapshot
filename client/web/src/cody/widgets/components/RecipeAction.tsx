import classNames from 'classnames'

import { MenuItem } from '@sourcegraph/wildcard'

import styles from './Recipes.module.scss'

export interface RecipeActionProps {
    title: string
    onClick: () => void
}

export const RecipeAction = ({ title, onClick }: RecipeActionProps): JSX.Element => (
    <MenuItem className={classNames(styles.recipeMenuWrapper)} onSelect={onClick}>
        {title}
    </MenuItem>
)
