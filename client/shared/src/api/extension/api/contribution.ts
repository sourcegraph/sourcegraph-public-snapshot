import { mapValues } from 'lodash'

import type { ContributableMenu, Contributions, Evaluated, MenuItemContribution, Raw } from '@sourcegraph/client-api'
import { type Context, Expression, parse, parseTemplate } from '@sourcegraph/template-parser'

/**
 * Merges the contributions.
 *
 * Most callers should use ContributionRegistry#getContributions, which merges all registered
 * contributions.
 */
export function mergeContributions(contributions: Evaluated<Contributions>[]): Evaluated<Contributions> {
    if (contributions.length === 0) {
        return {}
    }
    if (contributions.length === 1) {
        return contributions[0]
    }
    const merged: Evaluated<Contributions> = {}
    for (const contribution of contributions) {
        // swallow errors from malformed manifests to prevent breaking other
        // contributions or extensions: https://github.com/sourcegraph/sourcegraph/pull/12573
        if (contribution.actions) {
            try {
                if (!merged.actions) {
                    merged.actions = [...contribution.actions]
                } else {
                    merged.actions = [...merged.actions, ...contribution.actions]
                }
            } catch {
                // noop
            }
        }
        if (contribution.menus) {
            try {
                if (!merged.menus) {
                    merged.menus = { ...contribution.menus }
                } else {
                    for (const [menu, items] of Object.entries(contribution.menus) as [
                        ContributableMenu,
                        Evaluated<MenuItemContribution>[]
                    ][]) {
                        const mergedItems = merged.menus[menu]
                        try {
                            if (!mergedItems) {
                                merged.menus[menu] = [...items]
                            } else {
                                merged.menus[menu] = [...mergedItems, ...items]
                            }
                        } catch {
                            // noop
                        }
                    }
                }
            } catch {
                // noop
            }
        }
        if (contribution.searchFilters) {
            try {
                if (!merged.searchFilters) {
                    merged.searchFilters = [...contribution.searchFilters]
                } else {
                    merged.searchFilters = [...merged.searchFilters, ...contribution.searchFilters]
                }
            } catch {
                // noop
            }
        }
    }
    return merged
}

/**
 * Filters the contributions to only those that are enabled in the current context.
 */
export function filterContributions(contributions: Evaluated<Contributions>): Evaluated<Contributions> {
    if (!contributions.menus) {
        return contributions
    }
    return {
        ...contributions,
        menus: mapValues(contributions.menus, menuItems => menuItems?.filter(menuItem => menuItem.when !== false)),
    }
}

/**
 * Evaluates expressions in contribution definitions against the given context.
 *
 * @todo could walk object recursively
 */
export function evaluateContributions<T>(context: Context<T>, contributions: Contributions): Evaluated<Contributions> {
    return {
        ...contributions,
        menus:
            contributions.menus &&
            mapValues(contributions.menus, (menuItems): Evaluated<MenuItemContribution>[] | undefined =>
                menuItems?.map(menuItem => ({
                    ...menuItem,
                    when: menuItem.when && !!menuItem.when.exec(context),
                    disabledWhen: menuItem.disabledWhen && !!menuItem.disabledWhen.exec(context),
                }))
            ),
        actions: evaluateActionContributions<T>(context, contributions.actions),
    }
}

/**
 * Evaluates expressions in contribution definitions against the given context.
 */
function evaluateActionContributions<T>(
    context: Context<T>,
    actions: Contributions['actions']
): Evaluated<Contributions['actions']> {
    return actions?.map(action => ({
        ...action,
        title: action.title?.exec(context),
        disabledTitle: action.disabledTitle?.exec(context),
        category: action.category?.exec(context),
        description: action.description?.exec(context),
        iconURL: action.iconURL?.exec(context),
        actionItem: action.actionItem && {
            ...action.actionItem,
            label: action.actionItem.label?.exec(context),
            description: action.actionItem.description?.exec(context),
            iconURL: action.actionItem.iconURL?.exec(context),
            iconDescription: action.actionItem.iconDescription?.exec(context),
            pressed: action.actionItem.pressed?.exec(context),
        },
        commandArguments: action.commandArguments?.map(argument =>
            argument instanceof Expression ? argument.exec(context) : argument
        ),
    }))
}

/**
 * Parses expressions in contribution defitions so that they can be evaluated against various contexts.
 */
export function parseContributionExpressions(contributions: Raw<Contributions>): Contributions {
    return {
        ...contributions,
        menus:
            contributions.menus &&
            mapValues(contributions.menus, (menuItems): MenuItemContribution[] | undefined =>
                menuItems?.map(menuItem => ({
                    ...menuItem,
                    when: typeof menuItem.when === 'string' ? parse<boolean>(menuItem.when) : undefined,
                    disabledWhen:
                        typeof menuItem.disabledWhen === 'string' ? parse<boolean>(menuItem.disabledWhen) : undefined,
                }))
            ),
        actions: contributions && parseActionContributionExpressions(contributions.actions),
    }
}

const maybe = <T, R>(value: T | undefined, function_: (value: T) => R): R | undefined =>
    value === undefined ? undefined : function_(value)

/**
 * Evaluates expressions in contribution definitions against the given context.
 */
function parseActionContributionExpressions(actions: Raw<Contributions['actions']>): Contributions['actions'] {
    return actions?.map(action => ({
        ...action,
        title: maybe(action.title, parseTemplate),
        disabledTitle: maybe(action.disabledTitle, parseTemplate),
        category: maybe(action.category, parseTemplate),
        description: maybe(action.description, parseTemplate),
        iconURL: maybe(action.iconURL, parseTemplate),
        actionItem: action.actionItem && {
            ...action.actionItem,
            label: maybe(action.actionItem.label, parseTemplate),
            description: maybe(action.actionItem.description, parseTemplate),
            iconURL: maybe(action.actionItem.iconURL, parseTemplate),
            iconDescription: maybe(action.actionItem.iconDescription, parseTemplate),
            pressed: maybe(action.actionItem.pressed, pressed => parse(pressed)),
        },
        commandArguments: action.commandArguments?.map(argument =>
            typeof argument === 'string' ? parseTemplate(argument) : argument
        ),
    }))
}
