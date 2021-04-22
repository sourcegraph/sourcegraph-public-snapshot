import { FormApi } from 'final-form';
import React, { useCallback } from 'react';

import { CreateInsightForm, CreateInsightFormFields } from './components/create-insight-form/CreateInsightForm';

export interface CreateInsightPageProps {}

export const CreateInsightPage: React.FunctionComponent<CreateInsightPageProps> = props => {

    const handleSubmit = useCallback(
        (values: CreateInsightFormFields, form: FormApi<CreateInsightFormFields>) => {
            console.log(values, form)
        },
        []
    );

    return (
        <CreateInsightForm
            onSubmit={handleSubmit}/>
    );
}
