import classnames from 'classnames';
import React from 'react'

import { PageHeader, Container } from '@sourcegraph/wildcard/src';

import { Page } from '../../../../components/Page';
import { PageTitle } from '../../../../components/PageTitle';
import { CodeInsightsIcon } from '../../../components';

import { InsightsDashboardCreationContent } from './components/insights-dashboard-creation-content/InsightsDashboardCreationContent';
import styles from './InsightsDashboardCreationPage.module.scss'

interface InsightsDashboardCreationPageProps {}

export const InsightsDashboardCreationPage: React.FunctionComponent<InsightsDashboardCreationPageProps> = props => {
    const {} = props

    return (
        <Page className={classnames('col-8', styles.page)}>
            <PageTitle title="Create new code insight" />

            <PageHeader path={[{ icon: CodeInsightsIcon, text: 'Add new dashboard' }]}/>

            <Container className='mt-4'>
                <InsightsDashboardCreationContent/>
            </Container>
        </Page>
    )
}
