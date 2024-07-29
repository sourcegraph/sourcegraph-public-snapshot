import { Status, type NavigationEntry, type NavigationMenu } from '$lib/navigation/mainNavigation'

export const enum Mode {
    ENTERPRISE = 1 << 0,
    DOTCOM = 1 << 1,
    CODY_INSTANCE_ENABLED = 1 << 2,
    CODY_USER_ENABLED = 1 << 3,
    BATCH_CHANGES_ENABLED = 1 << 4,
    CODE_INSIGHTS_ENABLED = 1 << 5,
    AUTHENTICATED = 1 << 6,
    UNAUTHENTICATED = 1 << 7,
}

interface NavigationEntryDefinition extends Omit<NavigationEntry, 'href'> {
    href: string | [Mode, string][] | null
    mode?: Mode
}

interface NavigationMenuDefinition extends Omit<NavigationMenu, 'children' | 'href'> {
    href: string | [Mode, string][] | null
    children: NavigationEntryDefinition[]
    mode?: Mode
}

function matchesMode(entry: NavigationEntryDefinition, mode: Mode): boolean {
    return entry.mode === undefined || (entry.mode & mode) === entry.mode
}

function toEntry(entry: NavigationEntryDefinition, mode: Mode): NavigationEntry {
    return {
        ...entry,
        href: matchHref(entry.href, mode),
    }
}

function matchHref(href: NavigationEntryDefinition['href'], mode: Mode): string {
    if (!href) {
        return ''
    }

    if (typeof href === 'string') {
        return href
    }

    for (const [key, value] of href) {
        if ((mode & +key) === +key) {
            return value
        }
    }
    return ''
}

export function getMainNavigationEntries(mode: Mode): (NavigationMenu | NavigationEntry)[] {
    return navigationEntries
        .filter(entry => matchesMode(entry, mode))
        .map(definition => {
            const entry = toEntry(definition, mode)
            return 'children' in definition
                ? {
                      ...entry,
                      children: definition.children
                          .filter(child => matchesMode(child, mode))
                          .map(child => toEntry(child, mode)),
                  }
                : entry
        })
}

const navigationEntries: (NavigationMenuDefinition | NavigationEntryDefinition)[] = [
    {
        label: 'Code Search',
        icon: ILucideSearch,
        href: '/search',
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
        mode: Mode.DOTCOM | Mode.UNAUTHENTICATED,
    },
    {
        label: 'Cody',
        icon: ISgCody,
        href: '/cody/chat',
        mode: Mode.DOTCOM | Mode.AUTHENTICATED,
    },
    {
        label: 'Cody',
        icon: ISgCody,
        href: [
            [Mode.CODY_USER_ENABLED, '/cody/chat'],
            [Mode.CODY_INSTANCE_ENABLED, '/cody/dashboard'],
        ],
        mode: Mode.ENTERPRISE | Mode.CODY_INSTANCE_ENABLED,
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
    {
        label: 'Tools',
        icon: IMdiTools,
        href: null,
        children: [
            {
                label: 'Saved Searches',
                href: '/saved-searches',
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
        mode: Mode.ENTERPRISE,
    },
]
