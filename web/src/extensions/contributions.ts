/**
 * Contributions describe the functionality provided by an extension.
 *
 * See the github.com/sourcegraph/sourcegraph/cxp package for canonical documentation.
 */

export interface Contributions {
    commands?: CommandContribution[]
    menus?: MenuContributions
}

export interface CommandContribution {
    command: string
    title?: string
    iconURL?: string
    experimentalSettingsAction?: CommandContributionSettingsAction
}

interface CommandContributionSettingsAction {
    path: (string | number)[]
    cycleValues?: any[]
    prompt?: string
}

export enum ContributableMenu {
    EditorTitle = 'editor/title',
}

interface MenuContributions extends Record<ContributableMenu, MenuItemContribution[]> {}

interface MenuItemContribution {
    command: string
    hidden?: boolean
}
