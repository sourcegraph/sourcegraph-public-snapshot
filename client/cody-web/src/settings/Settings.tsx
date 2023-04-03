import React, { useCallback, useState } from 'react'

import { isEqual } from 'lodash'

import styles from './Settings.module.css'
import { WebConfiguration, useConfig } from './useConfig'

const SAMPLE_PUBLIC_CODEBASES = ['github.com/sourcegraph/sourcegraph', 'github.com/hashicorp/errwrap']

export const Settings: React.FunctionComponent<{
    config: WebConfiguration
    setConfig: ReturnType<typeof useConfig>[1]
}> = ({ config, setConfig }) => {
    const [pendingConfig, setPendingConfig] = useState<WebConfiguration>()
    const onInput = useCallback<React.FormEventHandler<HTMLInputElement>>(
        event => {
            const { name, value } = event.currentTarget
            setPendingConfig(prev => {
                const base = prev ?? config
                const updated: WebConfiguration = {
                    ...base,
                    [name]: value,
                }
                if (isEqual(updated, config)) {
                    return undefined // no changes vs. applied config
                }
                return updated
            })
        },
        [config]
    )

    const onApply = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            if (pendingConfig) {
                setConfig(pendingConfig)
            }
            setPendingConfig(undefined)
        },
        [pendingConfig, setConfig]
    )

    const sampleCodebases = config.serverEndpoint === 'https://sourcegraph.com' ? SAMPLE_PUBLIC_CODEBASES : null

    return (
        <aside className={styles.container}>
            <form className={styles.form} onSubmit={onApply}>
                <label className={styles.label}>
                    Sourcegraph URL{' '}
                    <input
                        name="serverEndpoint"
                        type="url"
                        required={true}
                        value={pendingConfig?.serverEndpoint ?? config.serverEndpoint}
                        onInput={onInput}
                        size={24}
                    />
                </label>
                <label className={styles.label}>
                    Access token{' '}
                    <input
                        name="accessToken"
                        type="password"
                        value={pendingConfig?.accessToken ?? config?.accessToken ?? ''}
                        onInput={onInput}
                        size={12}
                    />
                </label>
                <label className={styles.label}>
                    Codebase{' '}
                    <input
                        name="codebase"
                        type="text"
                        value={pendingConfig?.codebase ?? config.codebase ?? ''}
                        onInput={onInput}
                        list="codebases"
                        size={30}
                    />
                    {sampleCodebases && (
                        <datalist id="codebases">
                            {sampleCodebases.map(codebase => (
                                <option key={codebase} value={codebase} />
                            ))}
                        </datalist>
                    )}
                </label>
                <button type="submit" className={styles.button} disabled={!pendingConfig}>
                    Apply
                </button>
            </form>
        </aside>
    )
}
