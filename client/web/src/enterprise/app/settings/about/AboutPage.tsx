import classNames from 'classnames'

import { Container, H3, PageHeader } from '@sourcegraph/wildcard'

import styles from './AboutPage.module.scss'

export const AboutTabTitle = 'About Cody App'

export const AboutTab: React.FC = () => (
    <div className={classNames(styles.root)}>
        <PageHeader headingElement="h2" path={[{ text: AboutTabTitle }]} actions={[]} className="mb-3" />
        <Container className="mb-3">Version: {window.context.version}</Container>
        <Container className="mb-3 pb-2">
            <H3>Terms and Conditions</H3>
            <ul>
                <li>
                    <a href="https://about.sourcegraph.com/terms/cody-notice">Usage and Privacy Notice</a>
                </li>
                <li>
                    <a href="https://about.sourcegraph.com/terms/cloud">Terms of Service for Sourcegraph Cloud</a>
                </li>
            </ul>
        </Container>
    </div>
)
