import * as React from 'react'
import { useCallback } from 'react'

import { mdiCompass } from '@mdi/js'

import { Button, Select, Text, H1, H4, Icon } from '@sourcegraph/wildcard'

// eslint-disable-next-line import/extensions
import * as notebooks from '../../../hackathon-idea/db/notebooks.json'
import * as allPackages from '../../../hackathon-idea/src/packages.json'
import { Page } from '../components/Page'

import styles from './PackageRemixArea.module.scss'

interface PackageRemixAreaProps {}

const sortFunction = (a: string, b: string) => a.localeCompare(b, 'en', { numeric: true })

const PACKAGE_NAMES = allPackages
    .slice(0, 100)
    .map(package_ => package_.name)
    .sort(sortFunction)

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
        const [a, b] = [packageA, packageB].sort(sortFunction)
        const URL = `/notebooks/${notebooks[a][b]}`
        location.href = URL
    }

    return (
        <Page className={styles.packageRemixArea}>
            <div style={{ marginBottom: '1rem' }}>
                <div style={{ display: 'flex', flexDirection: 'row', alignItems: 'baseline' }}>
                    <Icon size="md" svgPath={mdiCompass} aria-label="explore" style={{ marginRight: '0.5rem' }} />
                    <H1> Package explorer </H1>
                </div>
                <Text> Learn how to use packages together using real-world examples </Text>
            </div>
            <H4> How do I use: </H4>
            <div style={{ display: 'flex', width: '80%', justifyContent: 'center', flexDirection: 'column' }}>
                <div
                    style={{
                        display: 'flex',
                        flexDirection: 'row',
                        alignItems: 'baseline',
                        justifyContent: 'space-between',
                        paddingTop: '2px',
                        width: '100%',
                    }}
                >
                    <Select aria-labelledby="package a" value={packageA} onChange={handlePackageASelect}>
                        {PACKAGE_NAMES.filter(pck => pck !== packageB).map(pack => (
                            <option key={pack} value={pack}>
                                {pack}
                            </option>
                        ))}
                    </Select>
                    <Text> with </Text>
                    <Select aria-labelledby="package b" value={packageB} onChange={handlePackageBSelect}>
                        {PACKAGE_NAMES.filter(pck => pck !== packageA).map(pack => (
                            <option key={pack} value={pack}>
                                {pack}
                            </option>
                        ))}
                    </Select>
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <Button
                        variant="primary"
                        as="a"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="mb-3"
                        onClick={redirectToNotebook}
                    >
                        Show me
                    </Button>
                </div>
            </div>
        </Page>
    )
}
