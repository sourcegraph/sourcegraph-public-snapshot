import { NavItemWithIconDescriptor } from '../../util/contributions'

export const enterpriseNamespaceAreaHeaderNavItems: readonly Pick<
    NavItemWithIconDescriptor,
    Exclude<keyof NavItemWithIconDescriptor, 'condition'>
>[] = []
