import { FunctionComponent, useCallback, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { gql, useLazyQuery, useMutation } from '@sourcegraph/http-client'
import { IFeatureFlagOverride } from '@sourcegraph/shared/src/schema'
import { Input, Alert, Typography } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import { Maybe, OrganizationVariables } from '../../graphql-operations'

const ORG_CODE_FLAG = 'org-code'

interface OrgResult {
    organization: Maybe<{
        id: string
    }>
}

interface FeatureFlagOverrideResult {
    createFeatureFlagOverride: Maybe<IFeatureFlagOverride>
}

interface FeatureFlagOverrideVariables {
    namespace: string
    flagName: string
    value: boolean
}

export const ORG_ID_BY_NAME = gql`
    query OrganizationIDByName($name: String!) {
        organization(name: $name) {
            id
        }
    }
`

export const CREATE_FEATURE_FLAG_OVERRIDE = gql`
    mutation CreateFeatureFlagOverride($namespace: ID!, $flagName: String!, $value: Boolean!) {
        createFeatureFlagOverride(namespace: $namespace, flagName: $flagName, value: $value) {
            __typename
        }
    }
`
/**
 * Form that sets a feature flag override for org-code flag, based on organization name.
 * This enables 2 screens - organization code host connections and organization repositories.
 *
 * On submit, there are 2 requests in series - first gets the organization ID by name,
 * second sets the feature flag override based on the ID from first request.
 *
 * This implementation is a quick hack for making our lives easier while in early access
 * stage. IMO it's not worth a lot of effort as it is throwaway work once we are in GA.
 */
export const EarlyAccessOrgsCodeForm: FunctionComponent<React.PropsWithChildren<any>> = () => {
    const [name, setName] = useState<string>('')

    const [updateFeatureFlag, { data, loading: flagLoading, error: flagError }] = useMutation<
        FeatureFlagOverrideResult,
        FeatureFlagOverrideVariables
    >(CREATE_FEATURE_FLAG_OVERRIDE, {
        onError: apolloError => console.error('Error when creating feature flag override', apolloError),
    })

    const [getOrgID, { loading: orgLoading, error: orgError }] = useLazyQuery<OrgResult, OrganizationVariables>(
        ORG_ID_BY_NAME,
        {
            onCompleted: ({ organization }) => {
                const id = organization?.id
                if (id) {
                    return updateFeatureFlag({
                        variables: {
                            namespace: id,
                            flagName: ORG_CODE_FLAG,
                            value: true,
                        },
                    })
                }
                return
            },
        }
    )

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            return getOrgID({
                variables: {
                    name,
                },
            })
        },
        [getOrgID, name]
    )

    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            setName(event.target.value)
            // clear the success message on input change
            if (data) {
                data.createFeatureFlagOverride = null
            }
        },
        [data]
    )

    const loading = orgLoading || flagLoading
    const error = orgError || flagError

    return (
        <Form onSubmit={onSubmit}>
            <Typography.H2>Organizations code early access</Typography.H2>
            <p>Type in an organization name to enable early access for organization code host and repositories.</p>

            <div className="d-flex justify-content-start align-items-end">
                <Input label="Name" value={name} onChange={onChange} className="mb-0" />
                <LoaderButton
                    className="ml-3"
                    type="submit"
                    loading={loading}
                    disabled={loading || name.trim().length < 3}
                    alwaysShowLabel={true}
                    label="Enable"
                    variant="primary"
                />
            </div>

            {error && (
                <Alert className="mt-3" variant="danger">
                    {error.message}
                </Alert>
            )}
            {data?.createFeatureFlagOverride && (
                <Alert className="mt-3 mb-0" variant="success">
                    Feature flag override created.
                </Alert>
            )}
        </Form>
    )
}
