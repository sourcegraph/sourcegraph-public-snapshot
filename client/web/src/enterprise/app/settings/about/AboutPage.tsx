import { mdiLaunch } from '@mdi/js'
import classNames from 'classnames'

import { Container, H1, H3, Icon, Link, PageHeader, Text } from '@sourcegraph/wildcard'

import { AboutPageUpdatePanel } from '../../../../cody/update/AboutPageUpdate'

import styles from './AboutPage.module.scss'

const assetsRoot = window.context?.assetsRoot || ''

export interface AboutTabProps {
    version?: string
}

export const AboutTab: React.FC<AboutTabProps> = ({ version = window.context.version }) => (
    <div className={classNames(styles.root)}>
        <PageHeader headingElement="h2" path={[{ text: 'About Cody' }]} actions={[]} className="mb-3" />
        <Container className="mb-3 p-0">
            <div className={classNames('d-flex align-items-center', styles.panel)}>
                <img
                    className={classNames('m-0', styles.icon)}
                    src={`${assetsRoot}/img/cody-logo-filled.png`}
                    alt="app logo"
                />
                {/* <CodyLogoFilled className={classNames('m-0', styles.icon)} /> */}
                <H1 className="m-0">Cody</H1>
            </div>
            <div className={classNames(styles.panel)}>
                <H3>Version</H3>
                <Text className="mb-1">{version}</Text>
                <AboutPageUpdatePanel />
            </div>
            <div className={classNames(styles.panel)}>
                <H3>Join Our Discord</H3>
                <Link to="https://discord.gg/rDPqBejz93">
                    Discord <Icon role="img" aria-hidden={true} svgPath={mdiLaunch} />
                </Link>
            </div>
            <div className={classNames(styles.panel)}>
                <H3>Terms and Conditions</H3>
                <div>
                    <Link to="https://about.sourcegraph.com/terms/cody-notice">
                        Usage and Privacy Notice <Icon role="img" aria-hidden={true} svgPath={mdiLaunch} />
                    </Link>
                </div>
                <div>
                    <Link to="https://about.sourcegraph.com/terms/cloud">
                        Terms of Service for Sourcegraph Cloud{' '}
                        <Icon role="img" aria-hidden={true} svgPath={mdiLaunch} />
                    </Link>
                </div>
            </div>
        </Container>
    </div>
)
