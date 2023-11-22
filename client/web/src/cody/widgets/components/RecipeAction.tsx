import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { AnchorLink, MenuItem, Icon, MenuLink } from '@sourcegraph/wildcard'

import styles from './Recipes.module.scss'

export type RecipeActionProps = {
    title: string
    disabled?: boolean
} & ({ onClick: () => void } | { to: string })

export const RecipeAction = ({ title, disabled, ...props }: RecipeActionProps): JSX.Element =>
    'onClick' in props ? (
        <MenuItem className={classNames(styles.recipeMenuWrapper)} onSelect={props.onClick} disabled={disabled}>
            {title}
        </MenuItem>
    ) : (
        <MenuLink as={AnchorLink} to={props.to} className={classNames(styles.recipeMenuWrapper)} disabled={disabled}>
            {title} <Icon aria-hidden={true} className="ml-1" svgPath={mdiOpenInNew} />
        </MenuLink>
    )
