import classnames from 'classnames';
import React from 'react'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors';
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings';
import { PageHeader, Container } from '@sourcegraph/wildcard/src';

import { AuthenticatedUser } from '../../../../auth';
import { Page } from '../../../../components/Page';
import { PageTitle } from '../../../../components/PageTitle';
import { Settings } from '../../../../schema/settings.schema';
import { CodeInsightsIcon } from '../../../components';

import { InsightsDashboardCreationContent } from './components/insights-dashboard-creation-content/InsightsDashboardCreationContent';
import styles from './InsightsDashboardCreationPage.module.scss'

const DEFAULT_FINAL_SETTINGS = {}

interface InsightsDashboardCreationPageProps extends SettingsCascadeProps<Settings> {

    authenticatedUser: AuthenticatedUser
}

export const InsightsDashboardCreationPage: React.FunctionComponent<InsightsDashboardCreationPageProps> = props => {
    const { authenticatedUser, settingsCascade } = props

    const handleSubmit = (): void => {

    }

    const handleCancel = (): void => {

    }

    const finalSettings = !isErrorLike(settingsCascade.final) && settingsCascade.final
        ? settingsCascade.final
        : DEFAULT_FINAL_SETTINGS

    return (
        <Page className={classnames('col-8', styles.page)}>
            <PageTitle title="Create new code insight" />

            <PageHeader path={[{ icon: CodeInsightsIcon, text: 'Add new dashboard' }]}/>

            <Container className='mt-4'>
                <InsightsDashboardCreationContent
                    settings={finalSettings}
                    organizations={authenticatedUser.organizations.nodes}
                    onSubmit={handleSubmit}
                    onCancel={handleCancel}
                />
            </Container>
        </Page>
    )
}
