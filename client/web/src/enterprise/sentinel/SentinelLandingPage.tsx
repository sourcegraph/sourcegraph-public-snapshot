import type { FC } from 'react'

import { mdiOpenInNew, mdiAccount, mdiCommentOutline } from '@mdi/js'
import ShieldHalfFullIcon from 'mdi-react/ShieldHalfFullIcon'
import { useNavigate } from 'react-router-dom'

import { Badge, Button, Card, CardBody, Icon, H2, H3, Text, PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'

import { LandingImage } from './LandingImage'

import styles from './SentinelLandingPage.module.scss'

export const SentinelLandingPage: FC = () => {
    const navigate = useNavigate()
    return (
        <Page>
            <div className={styles.headerContainer}>
                <PageHeader path={[{ icon: ShieldHalfFullIcon, text: 'Sentinel' }]} className={styles.header} />
                <div className={styles.badge}>
                    <Badge variant="info">Coming Soon</Badge>
                </div>
                <div className={styles.ctaBtn}>
                    <Button variant="primary" onClick={() => navigate('/sentinel/demo')}>
                        <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} /> Try Sentinel
                    </Button>
                </div>
            </div>
            <div>
                <Card>
                    <CardBody className={styles.container}>
                        <div className={styles.textContainer}>
                            <H2>Monitor, prioritise and fix vulnerabilities</H2>
                            <Text className="pb-3">
                                Sentinel is a Software Composition Analysis (SCA) tool that lets you monitor, prioritise
                                and fix vulnerable open source dependencies.
                            </Text>
                            <H3>Cut down the noise</H3>
                            <ul className="pb-3">
                                <li>Filter only reachable dependencies</li>
                                <li>Prioritise, accept and discard alerts </li>
                                <li>Narrow your search by repo, language or vulnerability severity</li>
                            </ul>
                            <H3>Fully integrated with Sourcegraph</H3>
                            <ul className="pb-2">
                                <li>Use Batch Changes to fix vulnerabilities across repos and code-hosts</li>
                                <li>Track the number of occurrences over time</li>
                                <li>Configure alerts for new vulnerabilities, based on severity or scope</li>
                                <li>Assign remediation to the best code owner</li>
                            </ul>
                        </div>
                        <div className={styles.buttonContainer}>
                            <Button
                                className={styles.tryBtn}
                                variant="primary"
                                onClick={() => navigate('/sentinel/demo')}
                            >
                                <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} /> Try Sentinel
                            </Button>
                            <Button className={styles.btn} variant="secondary">
                                <Icon role="img" aria-label="Open in a new tab" svgPath={mdiAccount} /> Talk with a
                                Customer Engineer
                            </Button>
                            <Button className={styles.btn} variant="secondary">
                                <Icon role="img" aria-label="Open in a new tab" svgPath={mdiCommentOutline} /> Join the
                                Public Discussion
                            </Button>
                        </div>
                        <div className={styles.imgContainer}>
                            <img src={LandingImage} alt="preview" />
                        </div>
                    </CardBody>
                </Card>
            </div>
        </Page>
    )
}
