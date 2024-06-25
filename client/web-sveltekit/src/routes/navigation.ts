import { isRepoRoute } from '$lib/navigation'
import { Status, isCurrent, type NavigationEntry, type NavigationMenu } from '$lib/navigation/mainNavigation'

export const enum Mode {
    ENTERPRISE = 1 << 0,
    DOTCOM = 1 << 1,
    CODY_ENABLED = 1 << 2,
    BATCH_CHANGES_ENABLED = 1 << 3,
    CODE_INSIGHTS_ENABLED = 1 << 4,
}

interface NavigationEntryWithMode extends NavigationEntry {
    mode: Mode
}

interface NavigationMenuWithMode extends NavigationMenu {
    mode: Mode
}

export function getMainNavigation(mode: Mode): (NavigationMenuWithMode | NavigationEntryWithMode)[] {
    return navigation.filter(entry => (entry.mode & mode) !== 0)
}

const navigation: (NavigationMenuWithMode | NavigationEntryWithMode)[] = [
    {
        label: 'Code Search',
        icon: ILucideSearch,
        href: '/search',
        children: [
            {
                label: 'Search Home',
                href: '/search',
            },
            {
                label: 'Contexts',
                href: '/contexts',
            },
            {
                label: 'Notebooks',
                href: '/notebooks',
            },
            {
                label: 'Monitoring',
                href: '/code-monitoring',
            },
            {
                label: 'Code Ownership',
                href: '/own',
            },
            {
                label: 'Search Jobs',
                href: '/search-jobs',
                status: Status.BETA,
            },
        ],
        isCurrent(this: NavigationMenu, page) {
            // This is a special case of the code search menu: It is marked as "current" if the
            // current page is a repository route.
            return isRepoRoute(page.route?.id) || this.children.some(entry => isCurrent(entry, page))
        },
        mode: Mode.ENTERPRISE,
    },
    {
        label: 'Code Search',
        icon: ILucideSearch,
        href: '/search',
        mode: Mode.DOTCOM,
    },
    {
        label: 'Cody',
        icon: ISgCody,
        href: '/cody',
        isCurrent(this: NavigationMenu, page) {
            return this.children.some(entry => isCurrent(entry, page))
        },
        children: [
            {
                label: 'Dashboard',
                href: '/cody',
            },
            {
                label: 'Web Chat',
                href: '/cody/chat',
            },
        ],
        mode: Mode.CODY_ENABLED,
    },
    {
        label: 'Batch Changes',
        icon: ISgBatchChanges,
        href: '/batch-changes',
        mode: Mode.BATCH_CHANGES_ENABLED,
    },
    {
        label: 'Insights',
        icon: ILucideBarChartBig,
        href: '/insights',
        mode: Mode.CODE_INSIGHTS_ENABLED,
    },
    {
        label: 'About Sourcegraph',
        href: '/',
        mode: Mode.DOTCOM,
    },
]
