import { ProjectIcon } from '../projects/icons'
import { NavItemWithIconDescriptor } from '../util/contributions'

export const namespaceAreaHeaderNavItems: readonly Pick<
    NavItemWithIconDescriptor,
    Exclude<keyof NavItemWithIconDescriptor, 'condition'>
>[] = [
    {
        to: '/projects',
        label: 'Projects',
        icon: ProjectIcon,
    },
]
