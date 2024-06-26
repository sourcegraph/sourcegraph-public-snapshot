import { isRepoRoute } from '$lib/navigation'
import { Status, isCurrent, type NavigationEntry, type NavigationMenu } from '$lib/navigation/mainNavigation'

/**
 * The main navigation of the application.
 */
export const mainNavigation: (NavigationMenu | NavigationEntry)[] = [
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
    },
    {
        label: 'Batch Changes',
        icon: ISgBatchChanges,
        href: '/batch-changes',
    },
    {
        label: 'Insights',
        icon: ILucideBarChartBig,
        href: '/insights',
    },
]

/**
 * The main navigation for sourcegraph.com
 */
export const dotcomMainNavigation: (NavigationMenu | NavigationEntry)[] = [
    {
        label: 'Code Search',
        icon: ILucideSearch,
        href: '/search',
    },
    {
        label: 'Cody',
        icon: ISgCody,
        href: '/cody',
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
        isCurrent(this: NavigationMenu, page) {
            return this.children.some(entry => isCurrent(entry, page))
        },
    },
    {
        label: 'About Sourcegraph',
        href: '/',
    },
]
