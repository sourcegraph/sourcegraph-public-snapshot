import { Button, Icon } from '@sourcegraph/wildcard'

import './Recipes.scss'

export interface RecipeProps {
    title: string
    icon?: string
    onClick?: () => void
}

export function Recipe({ title = 'Undefined', icon, onClick }: RecipeProps) {
    return (
        <Button variant="secondary" outline={true} className="recipeWrapper" onClick={onClick}>
            {icon && <Icon aria-hidden={true} svgPath={icon} />} {title}
        </Button>
    )
}
