import { mdiLaunch } from '@mdi/js'
import classNames from 'classnames'

import { Container, H1, H3, Icon, Link, PageHeader, Text } from '@sourcegraph/wildcard'

import logo from '../../../../../../../src-tauri/icons/icon.png'

import styles from './AboutPage.module.scss'

interface AboutTabProps {}

export const AboutTab: React.FC<AboutTabProps> = () => (
    <div className={classNames(styles.root)}>
        <PageHeader headingElement="h2" path={[{ text: 'About Cody' }]} actions={[]} className="mb-3" />
        <Container className="mb-3 p-0">
            <div className={classNames('d-flex align-items-center', styles.panel)}>
                <img className={classNames('m-0', styles.icon)} src={logo as string} alt="app logo" />
                <H1 className="m-0">Cody</H1>
            </div>
            <div className={classNames(styles.panel)}>
                <H3>Version</H3>
                {window.context.version}

                {/* TODO(nelsona): Replace with real updates from 
                https://github.com/sourcegraph/sourcegraph/pull/54507 */}

                <Text className="mt-4 mb-0">
                    We're making regular improvements to the Cody app.
                    <br /> For information on how to upgrade to the latest version, see{' '}
                    <Link to="/help/app#upgrading" target="_blank" rel="noopener">
                        our docs
                    </Link>
                    .
                </Text>
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
