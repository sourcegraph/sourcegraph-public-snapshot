import { FORM_ERROR } from 'final-form';
import React, { useCallback, useContext } from 'react';
import { Redirect } from 'react-router';
import { RouteComponentProps } from 'react-router-dom';
import * as uuid from 'uuid'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context';
import { asError } from '@sourcegraph/shared/src/util/errors'

import { AuthenticatedUser } from '../../../auth';
import { InsightsApiContext } from '../../core/backend/api-provider'

import { CreateInsightForm, CreateInsightFormProps } from './components/create-insight-form/CreateInsightForm';

export interface CreateInsightPageProps extends PlatformContextProps, RouteComponentProps  {
    authenticatedUser: AuthenticatedUser | null
}

export const CreateInsightPage: React.FunctionComponent<CreateInsightPageProps> = props => {
    const { platformContext, authenticatedUser, history } = props;
    const { updateSubjectSettings, getSubjectSettings } = useContext(InsightsApiContext);

    const handleSubmit = useCallback<CreateInsightFormProps['onSubmit']>(
        async values => {

            if (!authenticatedUser) {
                return
            }

            const { id: userID, organizations: { nodes: orgs } } = authenticatedUser;
            const subjectID = values.visibility === 'personal'
                ? userID
                // TODO [VK] Add orgs picker in creation UI and not just pick first organization
                : orgs[0].id;

            try {
                const settings = await getSubjectSettings(subjectID).toPromise();
                const content = JSON.parse(settings.contents) as object;
                const insightID = uuid.v4();

                const newSettings = {
                    ...content,
                    [`searchInsights.insight.${insightID}`]: {
                        title: values.title,
                        repositories: values.repositories.split(','),
                        series: values.series.map(line => ({
                            name: line.name,
                            query: line.query,
                            stroke: line.color
                        })),
                        step: {
                            [values.step]: values.stepValue
                        }
                    }
                }

                await updateSubjectSettings(platformContext, subjectID, JSON.stringify(newSettings)).toPromise()

                history.push('/insights');
            } catch (error) {
                return { [FORM_ERROR]: asError(error)}
            }

            return;
        },
        [history, updateSubjectSettings, getSubjectSettings, platformContext, authenticatedUser]
    );

    if (authenticatedUser === null) {
        return <Redirect to='/'/>
    }

    return (
        <CreateInsightForm
            onSubmit={handleSubmit}/>
    );
}
