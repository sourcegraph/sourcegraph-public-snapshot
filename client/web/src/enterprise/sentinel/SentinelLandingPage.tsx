import { FC } from 'react'
import { useNavigate } from 'react-router-dom'
import ShieldHalfFullIcon from 'mdi-react/ShieldHalfFullIcon'
import { mdiOpenInNew, mdiAccount, mdiCommentOutline } from '@mdi/js'
import { Badge, Button, Card, CardBody, Icon, H2, H3, Text } from '@sourcegraph/wildcard'
import { PageHeader } from '@sourcegraph/wildcard'
import { Page } from '../../components/Page'

import styles from './SentinelLandingPage.module.scss'

export const SentinelLandingPage: FC = () => {
    const navigate = useNavigate()
    return (
        <Page>
            <div className={styles.headerContainer}>
                <PageHeader path={[{ icon: ShieldHalfFullIcon, text: 'Sentinel' }]} className={styles.header} />
                <div className={styles.badge}>
                    <Badge variant="info">{'Coming Soon'}</Badge>
                </div>
                <div className="text-right">
                    <Button variant="primary" onClick={() => navigate('/sentinel/demo')}>
                        <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} /> Try Sentinel
                    </Button>
                </div>
            </div>
            <div>
                <Card>
                    <CardBody>
                        <div>
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
                                <li>Assign remediation to the best owner with Sourcegraph Own</li>
                            </ul>
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
                    </CardBody>
                </Card>
            </div>
        </Page>
    )
}
