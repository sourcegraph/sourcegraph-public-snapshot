import * as OmniCLI from 'omnicli'

import storage from '../../extension/storage'
import { featureFlagDefaults, FeatureFlags } from '../../extension/types'
import * as featureFlags from '../util/featureFlags'

const keyIsFeatureFlag = (key: string): key is keyof FeatureFlags => featureFlagDefaults[key] !== undefined

function featureFlagAction([key, value]: string[]): void {
    if (!keyIsFeatureFlag(key)) {
        return
    }

    if (typeof featureFlagDefaults[key] === 'boolean') {
        featureFlags
            .set(key, JSON.parse(value))
            .then(() => {
                /*noop*/
            })
            .catch(err => console.log('unable to set feature flag'))
        return
    }

    // TODO: Support other types when we add flags with other types
}

const getFeatureFlagSuggestsions = (flagType?: 'boolean') => ([cmd, ...args]: string[]): Promise<
    OmniCLI.Suggestion[]
> =>
    new Promise(resolve => {
        storage.getSync(({ featureFlags }) => {
            const suggestions: OmniCLI.Suggestion[] = Object.keys(featureFlags)
                .filter(flag => (flagType ? typeof featureFlagDefaults[flag] === flagType : true))
                .map(flag => ({
                    content: flag,
                    description: `${flag} - ${typeof featureFlags[flag]}`,
                }))

            resolve(suggestions)
        })
    })

export const featureFlagsCommand: OmniCLI.Command = {
    name: 'feature-flag',
    alias: ['flag', 'ff'],
    action: featureFlagAction,
    getSuggestions: getFeatureFlagSuggestsions(),
    description: 'Set experimental feature flags',
}

function toggleFeatureFlagAction([key, value]: string[]): void {
    if (!keyIsFeatureFlag(key)) {
        return
    }

    featureFlags
        .toggle(key)
        .then(() => {
            /*noop*/
        })
        .catch(err => console.log('unable to set feature flag'))
}

export const toggleFeatureFlagsCommand: OmniCLI.Command = {
    name: 'toggle-feature-flag',
    alias: ['toggle-flag', 'tff'],
    action: toggleFeatureFlagAction,
    getSuggestions: getFeatureFlagSuggestsions('boolean'),
    description: 'Toggle an experimental feature flag',
}
