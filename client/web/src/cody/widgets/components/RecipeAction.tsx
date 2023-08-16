import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { AnchorLink, MenuItem, Icon, MenuLink } from '@sourcegraph/wildcard'

import styles from './Recipes.module.scss'

export interface RecipeActionProps {
    title: string
    onClick?: () => void
    to?: string
    disabled?: boolean
}

export const RecipeAction = ({ title, onClick, to, disabled }: RecipeActionProps): JSX.Element => (
    <>
        {!!onClick ? (
            <MenuItem className={classNames(styles.recipeMenuWrapper)} onSelect={onClick} disabled={disabled}>
                {title}
            </MenuItem>
        ) : !!to ? (
            <MenuLink as={AnchorLink} to={to}>
                {title} <Icon aria-hidden={true} className="ml-1" svgPath={mdiOpenInNew} />
            </MenuLink>
        ) : null}
    </>
)
