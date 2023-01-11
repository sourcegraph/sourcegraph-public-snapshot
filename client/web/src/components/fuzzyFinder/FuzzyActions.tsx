import { MutableRefObject } from 'react'

import { mdiCodeGreaterThan } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'

import { ThemePreference, ThemeState } from '../../theme'

import { FuzzyFSM, newFuzzyFSMFromValues } from './FuzzyFsm'

export class FuzzyAction {
    constructor(public readonly id: string, public readonly title: string, public readonly run: () => void) {}
}

export interface FuzzyActionProps {
    themeState: MutableRefObject<ThemeState>
}
export function getAllFuzzyActions(props: FuzzyActionProps): FuzzyAction[] {
    return [
        new FuzzyAction('toggle.theme', 'Toggle Between Dark/Light Theme', () => {
            const themeState = props.themeState.current
            switch (themeState.enhancedThemePreference) {
                case ThemePreference.Dark:
                    return themeState.setThemePreference(ThemePreference.Light)
                case ThemePreference.Light:
                    return themeState.setThemePreference(ThemePreference.Dark)
            }
        }),
    ]
}

export function createActionsFSM(actions: FuzzyAction[]): FuzzyFSM {
    return newFuzzyFSMFromValues(
        actions.map(action => ({
            text: action.title,
            onClick: action.run,
            icon: <Icon aria-label={action.title} svgPath={mdiCodeGreaterThan} />,
        })),
        undefined
    )
}
