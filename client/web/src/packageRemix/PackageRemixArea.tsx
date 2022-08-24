import * as React from 'react'
import { useCallback } from 'react'

import { Button, H1, Select } from '@sourcegraph/wildcard'

// eslint-disable-next-line import/extensions
import * as notebooks from '../../../hackathon-idea/db/notebooks.json'
import * as allPackages from '../../../hackathon-idea/src/packages.json'
import { Page } from '../components/Page'

import styles from './PackageRemixArea.module.scss'

interface PackageRemixAreaProps {}

const sortFn = (a: string, b: string) => a.localeCompare(b, 'en', { numeric: true })

let PACKAGE_NAMES = allPackages
    .slice(0, 100)
    .map(pkg => pkg.name)
    .sort(sortFn)

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
        const [a, b] = [packageA, packageB].sort(sortFn)
        console.log(a, b)
        const URL = `/notebooks/${notebooks[a][b]}`
        location.href = URL
    }

    return (
        <Page className={styles.packageRemixArea}>
            <H1> Package Remix</H1>
            <div style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-around', width: '100%' }}>
                <Select aria-labelledby="package a" value={packageA} onChange={handlePackageASelect}>
                    {PACKAGE_NAMES.filter(pck => pck !== packageB).map(pack => (
                        <option key={pack} value={pack}>
                            {pack}
                        </option>
                    ))}
                </Select>
                <Select aria-labelledby="package b" value={packageB} onChange={handlePackageBSelect}>
                    {PACKAGE_NAMES.filter(pck => pck !== packageA).map(pack => (
                        <option key={pack} value={pack}>
                            {pack}
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
