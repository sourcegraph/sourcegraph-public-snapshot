import { mdiCodeGreaterThan } from '@mdi/js'

import { Theme, ThemeSetting } from '@sourcegraph/shared/src/theme'
import { Icon } from '@sourcegraph/wildcard'

import { type FuzzyFSM, newFuzzyFSMFromValues } from './FuzzyFsm'

export class FuzzyAction {
    constructor(public readonly id: string, public readonly title: string, public readonly run: () => void) {}
}

export interface FuzzyActionProps {
    theme: Theme
    setThemeSetting: (theme: ThemeSetting) => void
}

export function getAllFuzzyActions(props: FuzzyActionProps): FuzzyAction[] {
    const { theme, setThemeSetting } = props

    return [
        new FuzzyAction('toggle.theme', 'Toggle Between Dark/Light Theme', () => {
            switch (theme) {
                case Theme.Dark: {
                    return setThemeSetting(ThemeSetting.Light)
                }
                case Theme.Light: {
                    return setThemeSetting(ThemeSetting.Dark)
                }
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
        }))
    )
}
