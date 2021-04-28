import React, { useCallback } from 'react'

import { Page } from '../../../../components/Page';
import { PageTitle } from '../../../../components/PageTitle';

import { LangStatsInsightCreationForm, LangStatsInsightCreationFormProps } from './components/lang-stats-insight-creation-form/LangStatsInsightCreationForm';
import styles from './CreationLangStatsInsightPage.module.scss'

export interface CreateLangStatsInsightPageProps {}

export const CreationLangStatsInsightPage: React.FunctionComponent<CreateLangStatsInsightPageProps> = props => {
    const handleSubmit = useCallback<LangStatsInsightCreationFormProps['onSubmit']>(values => {
        console.log('submit lang stats creation form', { values });
    }, []);

    return (
        <Page className='col-8'>
            <PageTitle title="Create new code insight" />

            <div className={styles.creationLangInsightPageSubTitleContainer}>
                <h2>Set up new language usage insight</h2>

                <p className="text-muted">
                    Shows usage of languages in your repository based on number of lines of code.{' '}
                    <a
                        href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                        target="_blank"
                        rel="noopener"
                    >
                        Learn more.
                    </a>
                </p>
            </div>

            <LangStatsInsightCreationForm onSubmit={handleSubmit}/>
        </Page>
    )
}
