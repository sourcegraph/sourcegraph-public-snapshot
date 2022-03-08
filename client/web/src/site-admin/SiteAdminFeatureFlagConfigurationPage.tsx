import classNames from 'classnames'
import React, { FunctionComponent, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps, useHistory } from 'react-router'
import { of } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Button, Container, Link, LoadingSpinner, PageHeader, Select, useObservable } from '@sourcegraph/wildcard'

import { Collapsible } from '../components/Collapsible'
import { RadioButtons } from '../components/RadioButtons'

import { fetchFeatureFlags as defaultFetchFeatureFlags } from './backend'
import styles from './SiteAdminFeatureFlagConfigurationPage.module.scss'
import { getFeatureFlagReferences, parseProductReference, Reference } from './SiteAdminFeatureFlagsPage'

export interface SiteAdminFeatureFlagConfigurationProps extends RouteComponentProps<{ name: string }>, TelemetryProps {
    fetchFeatureFlags?: typeof defaultFetchFeatureFlags
    productVersion?: string
}

export const SiteAdminFeatureFlagConfigurationPage: FunctionComponent<SiteAdminFeatureFlagConfigurationProps> = ({
    match: {
        params: { name },
    },
    fetchFeatureFlags = defaultFetchFeatureFlags,
    productVersion = window.context.version,
}) => {
    const history = useHistory()
    const productGitVersion = parseProductReference(productVersion)
    const featureFlagOrError = useObservable(
        useMemo(
            () =>
                name !== 'new'
                    ? fetchFeatureFlags().pipe(
                          map(flags => flags.find(flag => flag.name === name)),
                          catchError((error): [ErrorLike] => [asError(error)])
                      )
                    : of(undefined),
            [name, fetchFeatureFlags]
        )
    )

    const [flagName, setFlagName] = useState<string>()
    const [flagType, setFlagType] = useState<FeatureFlagType>()
    const [flagValue, setFlagValue] = useState<FeatureFlagValue>()
    const [flagOverrides, setOverrides] = useState<FeatureFlagOverride[]>()
    useEffect(() => {
        if (featureFlagOrError && !isErrorLike(featureFlagOrError)) {
            setFlagName(featureFlagOrError.name)
            setFlagType(featureFlagOrError.__typename)
            setFlagValue(featureFlagOrError)
            setOverrides(featureFlagOrError.overrides)
        }
    }, [featureFlagOrError])

    const references = useObservable(
        useMemo(() => (flagName ? getFeatureFlagReferences(flagName, productGitVersion) : of([])), [
            flagName,
            productGitVersion,
        ])
    )

    let body: React.ReactElement
    let actions: React.ReactElement | undefined
    if (name === 'new') {
        body = (
            <CreateFeatureFlag
                name={flagName}
                type={flagType}
                value={flagValue}
                setFlagName={setFlagName}
                setFlagType={setFlagType}
                setFlagValue={setFlagValue}
            />
        )
        actions = (
            <Button variant="primary" disabled={true}>
                Create feature flag
            </Button>
        )
    } else if (isErrorLike(featureFlagOrError)) {
        body = <ErrorAlert prefix="Error fetching feature flag policy" error={featureFlagOrError} />
    } else if (flagName && flagType && flagValue) {
        body = (
            <ManageFeatureFlag
                name={flagName}
                type={flagType}
                value={flagValue}
                overrides={flagOverrides}
                references={references}
                setFlagValue={setFlagValue}
            />
        )
        actions = (
            <>
                <Button className="mr-2" variant="primary" disabled={true}>
                    Update feature flag
                </Button>
                <Button variant="danger" disabled={true}>
                    Delete feature flag
                </Button>
            </>
        )
    } else {
        body = <LoadingSpinner className="mt-2" />
    }

    const verb = name === 'new' ? 'Create' : 'Manage'
    return (
        <>
            <PageTitle title={`${verb} feature flag`} />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>{verb} feature flag</>,
                    },
                ]}
                className="mb-3"
            />

            <Container>{body}</Container>

            {actions && (
                <div className="mt-3">
                    {actions}
                    <Button type="button" className="ml-2" variant="secondary" onClick={() => history.push('../')}>
                        Cancel
                    </Button>
                </div>
            )}
        </>
    )
}

type FeatureFlagType = 'FeatureFlagBoolean' | 'FeatureFlagRollout'

interface FeatureFlagOverride {
    id: string
    value: boolean
}

interface FeatureFlagBooleanValue {
    value: boolean
}

interface FeatureFlagRolloutValue {
    rolloutBasisPoints: number
}

type FeatureFlagValue = FeatureFlagBooleanValue | FeatureFlagRolloutValue

const ManageFeatureFlag: FunctionComponent<{
    name: string
    type: FeatureFlagType
    value: FeatureFlagValue
    overrides?: FeatureFlagOverride[]
    references?: Reference[]
    setFlagValue: (flag: FeatureFlagValue) => void
}> = ({ name, type, value, overrides, references, setFlagValue }) => (
    <>
        <h3>Name</h3>
        <p>{name}</p>

        <h3>Type</h3>
        <p>{type.slice('FeatureFlag'.length)}</p>

        <FeatureFlagValueSettings type={type} value={value} setFlagValue={setFlagValue} />

        <Collapsible
            title={<h3>Overrides</h3>}
            detail={`${overrides?.length || 0} ${overrides?.length !== 1 ? 'overrides' : 'override'}`}
            className="p-0 font-weight-normal"
            buttonClassName="mb-0"
            titleAtStart={true}
            defaultExpanded={false}
        >
            <div className={classNames('pt-2', styles.nodeGrid)}>
                {overrides?.map(override => (
                    <React.Fragment key={override.id}>
                        <div className="py-1 pr-2">
                            <code>{JSON.stringify(override.value)}</code>
                        </div>

                        <span className={classNames('py-1 pl-2', styles.nodeGridCode)}>
                            {/*
                                            TODO: querying for namespace connection seems to
                                            error out often, so just present the ID for now.
                                            https://github.com/sourcegraph/sourcegraph/issues/32238
                                        */}
                            <code>{override.id}</code>
                        </span>
                    </React.Fragment>
                ))}
            </div>
        </Collapsible>

        {references && references.length > 0 && (
            <>
                <br />
                <Collapsible
                    title={<h3>References</h3>}
                    detail={`${references.length} ${references.length > 1 ? 'references' : 'reference'}`}
                    className="p-0 font-weight-normal"
                    buttonClassName="mb-0"
                    titleAtStart={true}
                    defaultExpanded={false}
                >
                    <div className="pt-2">
                        {references.map(reference => (
                            <div key={name + reference.file}>
                                <Link target="_blank" rel="noopener noreferrer" to={reference.searchURL}>
                                    <code>{reference.file}</code>
                                </Link>
                            </div>
                        ))}
                    </div>
                </Collapsible>
            </>
        )}
    </>
)

const CreateFeatureFlag: React.FunctionComponent<{
    name?: string
    setFlagName: (s: string) => void
    type?: FeatureFlagType
    setFlagType: (t: FeatureFlagType) => void
    value?: FeatureFlagValue
    setFlagValue: (v: FeatureFlagValue) => void
}> = ({ name, setFlagName, type, setFlagType, value, setFlagValue }) => (
    <>
        <div className="form-group d-flex flex-column">
            <label htmlFor="name">
                <h3>Name</h3>
            </label>
            <input
                id="name"
                type="text"
                className="form-control"
                value={name}
                onChange={({ target: { value } }) => {
                    setFlagName(value)
                }}
            />
            <small className="form-text text-muted">Required.</small>
        </div>

        <Select
            id="type"
            label="Type"
            value={type}
            onChange={({ target: { value } }) => setFlagType(value as FeatureFlagType)}
            message="Required."
        >
            <option value="">Select flag type</option>
            <option value="FeatureFlagRollout">Rollout</option>
            <option value="FeatureFlagBoolean">Boolean</option>
        </Select>

        {type && <FeatureFlagValueSettings type={type} value={value} setFlagValue={setFlagValue} />}
    </>
)

const FeatureFlagValueSettings: React.FunctionComponent<{
    type: FeatureFlagType
    value?: FeatureFlagValue
    setFlagValue: (next: FeatureFlagValue) => void
}> = ({ type, value, setFlagValue }) => {
    if (type === 'FeatureFlagRollout') {
        return (
            <FeatureFlagRolloutValueSettings
                value={(value as FeatureFlagRolloutValue) || { rolloutBasisPoints: 0 }}
                update={next => {
                    setFlagValue({
                        ...value,
                        ...next,
                    })
                }}
            />
        )
    }

    return (
        <FeatureFlagBooleanValueSettings
            value={(value as FeatureFlagBooleanValue) || { value: 0 }}
            update={next => {
                setFlagValue({
                    ...value,
                    ...next,
                })
            }}
        />
    )
}

const FeatureFlagRolloutValueSettings: React.FunctionComponent<{
    value: FeatureFlagRolloutValue
    update: (next: FeatureFlagRolloutValue) => void
}> = ({ value, update }) => (
    <div className="form-group d-flex flex-column">
        <label htmlFor="rollout-value">
            <h3>Value</h3>
        </label>
        <output>
            {value.rolloutBasisPoints} basis points ({Math.floor(value.rolloutBasisPoints / 100)}%)
        </output>
        <input
            type="range"
            id="rollout-value"
            name="rollout-value"
            step="10"
            min="0"
            max="10000"
            className="w-50"
            value={value.rolloutBasisPoints}
            onChange={({ target }) => {
                update({ rolloutBasisPoints: parseInt(target.value, 10) })
            }}
        />
        <small className="form-text text-muted">Required.</small>
    </div>
)

const FeatureFlagBooleanValueSettings: React.FunctionComponent<{
    value: FeatureFlagBooleanValue
    update: (next: FeatureFlagBooleanValue) => void
}> = ({ value, update }) => {
    const radioButtons = [
        {
            id: 'true',
            label: 'True',
            tooltip: 'Enable this feature flag.',
        },
        {
            id: 'false',
            label: 'False',
            tooltip: 'Disable this feature flag.',
        },
    ]
    return (
        <div className="form-group d-flex flex-column">
            <label htmlFor="bool-value">
                <h3>Value</h3>
            </label>
            <RadioButtons
                nodes={radioButtons}
                name="bool-value"
                className="pt-0"
                selected={value.value ? 'true' : 'false'}
                onChange={({ target }) => {
                    update({ value: target.value === 'true' })
                }}
            />
            <small className="form-text text-muted">Required.</small>
        </div>
    )
}
