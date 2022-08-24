import * as React from 'react'
import { useCallback } from 'react'

import { Button, H1, Select } from '@sourcegraph/wildcard'

// eslint-disable-next-line import/extensions
import * as notebooks from '../../../hackathon-idea/db/notebooks.json'
import * as allPackages from '../../../hackathon-idea/src/packages.json'
import { Page } from '../components/Page'

import styles from './PackageRemixArea.module.scss'

interface PackageRemixAreaProps {}

export const PackageRemixArea: React.FunctionComponent<React.PropsWithChildren<PackageRemixAreaProps>> = () => {
    const [packageA, setPackageA] = React.useState('react')
    const [packageB, setPackageB] = React.useState('redux')
    const handlePackageASelect = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(event => {
        setPackageA(event.target.value)
    }, [])

    const handlePackageBSelect = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(event => {
        setPackageB(event.target.value)
    }, [])

    const redirectToNotebook = () => {
        const [a, b] = [packageA, packageB].sort((a, b) => a.localeCompare(b, 'en', { numeric: true }))
        const URL = `/notebooks/${notebooks[a][b]}`
        location.href = URL
    }

    return (
        <Page className={styles.packageRemixArea}>
            <H1> Package Remix</H1>
            <div style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-around', width: '100%' }}>
                <Select aria-labelledby="package a" value={packageA} onChange={handlePackageASelect}>
                    {allPackages
                        .filter(pck => pck.name !== packageB)
                        .map(pack => (
                            <option key={pack.name} value={pack.name}>
                                {pack.name}
                            </option>
                        ))}
                </Select>
                <Select aria-labelledby="package b" value={packageB} onChange={handlePackageBSelect}>
                    {allPackages
                        .filter(pck => pck.name !== packageA)
                        .map(pack => (
                            <option key={pack.name} value={pack.name}>
                                {pack.name}
                            </option>
                        ))}
                </Select>
                <Button
                    variant="primary"
                    as="a"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="mb-3"
                    onClick={redirectToNotebook}
                >
                    Show
                </Button>
            </div>
        </Page>
    )
}
