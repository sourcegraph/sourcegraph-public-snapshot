import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { MenuItem, Icon } from '@sourcegraph/wildcard'

import styles from './Recipes.module.scss'

export interface RecipeActionProps {
    title: string
    onClick: () => void
    disabled?: boolean
}

export const RecipeAction = ({ title, onClick, disabled }: RecipeActionProps): JSX.Element => (
    <MenuItem className={classNames(styles.recipeMenuWrapper)} onSelect={onClick} disabled={disabled}>
        {title}
        {title === 'Use locally' && <Icon svgPath={mdiOpenInNew} className="ml-1" />}
    </MenuItem>
)
