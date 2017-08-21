import { Classes as AutocompleteClasses } from '@sourcegraph/components/src/Autocomplete/style';
import * as React from 'react';
import { fetchRepos } from 'sourcegraph/backend';

export const defaultSearchGroups = [{ uri: 'active' }, { uri: 'inactive' }];

export interface RepoResult {
    description: string;
    fork: boolean;
    private: boolean;
    pushedAt: string;
    uri: string;
}

export function RepoResult(props: { highlighted: boolean, item: RepoResult, classes: AutocompleteClasses }): JSX.Element | null {
    return <div className={props.highlighted ? props.classes.itemSelected : props.classes.item}>
        {props.item.uri}
    </div>;
}

export function resolveRepos(query: string): Promise<any[]> {
    query = query.toLowerCase();
    if (query === '') {
        return Promise.resolve(defaultSearchGroups);
    }
    return fetchRepos(query);
}
