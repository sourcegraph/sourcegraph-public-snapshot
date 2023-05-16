import React, {FC, useState} from 'react';
import {PageTitle} from '../../../components/PageTitle';
import {
    Container,
    PageHeader,
    H3, Text, Label, Button, LoadingSpinner, ErrorAlert
} from '@sourcegraph/wildcard'
import './own-status-page-styles.scss'
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import {RepositoryPatternList} from '../../codeintel/configuration/components/RepositoryPatternList';
import {useMutation, useQuery} from '@sourcegraph/http-client';
import {
    GetOwnSignalConfigurationsResult, OwnSignalConfig,
    UpdateSignalConfigsResult,
    UpdateSignalConfigsVariables,
} from '../../../graphql-operations';
import {GET_OWN_JOB_CONFIGURATIONS, UPDATE_SIGNAL_CONFIGURATIONS} from './query';
import {noop} from 'rxjs';

export const OwnStatusPage: FC = () => {
    const [hasLocalChanges, setHasLocalChanges] = useState<boolean>(false)
    const [localData, setLocalData] = useState<OwnSignalConfig[]>([])
    const [saveError, setSaveError] = useState<Error>()

    const { loading, error } = useQuery<GetOwnSignalConfigurationsResult>(
        GET_OWN_JOB_CONFIGURATIONS, {onCompleted: data => {setLocalData(data.ownSignalConfigurations)}}
    )

    const [saveConfigs, {loading: loadingSaveConfigs}] = useMutation<UpdateSignalConfigsResult, UpdateSignalConfigsVariables>(UPDATE_SIGNAL_CONFIGURATIONS, {})

    function onUpdateJob(index: number, newJob: OwnSignalConfig): void {
        setHasLocalChanges(true)
        const newData = localData.map((job: OwnSignalConfig, ind: number) => {
            if (ind === index) {
                return newJob
            }
            return job
        })
        setLocalData(newData)
    }

    return (
    <div>
        <span className='topHeader'>
            <div>
                <PageTitle title="Own Signals Configuration"/>
                <PageHeader
                    headingElement="h2"
                    path={[{text: 'Own Signals Configuration'}]}
                    description="List of Own inference signal indexers and their configurations. All repositories are included by default."
                    className="mb-3"
                />
                {saveError && <ErrorAlert error={saveError}/>}
            </div>

            {<Button id='saveButton' disabled={!hasLocalChanges} aria-label="Save changes" variant="primary" onClick={() => {
                setSaveError(null)
                // do network stuff
                saveConfigs({variables: {
                    input: {
                        configs: localData.map(ldd => ({name: ldd.name, enabled: ldd.isEnabled, excludedRepoPatterns: ldd.excludedRepoPatterns}))
                    }
                    }}).then(result => {
                        if (result.errors) {
                            setSaveError(new Error('Failed to save configurations'))
                        } else {
                            setHasLocalChanges(false)
                            setLocalData(result.data.updateOwnSignalConfigurations)
                        }
                }).catch(noop)
            }}>
                {loadingSaveConfigs && <LoadingSpinner/>}
                {!loadingSaveConfigs && 'Save'}
            </Button>}
        </span>

        <Container className='root'>
            {(loading) && <LoadingSpinner/>}
            {error && <ErrorAlert prefix="Error fetching Own signal configurations" error={error} /> }
            {!(loading) && localData && !error && localData.map((job: OwnSignalConfig, index: number) => (
                <li key={job.name} className="job">
                    <div className='jobHeader'>
                        <H3 className='jobName'>{job.name}</H3>
                        <div id="job-item" className='jobStatus'>
                            <Toggle
                                onToggle={value => {
                                    onUpdateJob(index, {...job, isEnabled: value})
                                }}
                                title={job.isEnabled ? 'Enabled' : 'Disabled'}
                                id="job-enabled"
                                value={job.isEnabled}
                                aria-label={`Toggle ${job.name} job`}
                            />
                            <Text id='statusText' size="small" className="text-muted mb-0">
                             {job.isEnabled ? 'Enabled' : 'Disabled'}
                            </Text>
                        </div>
                    </div>
                    <span className='jobDescription'>{job.description}</span>

                    <div id='excludeRepos'>
                        <Label className="mb-0">Exclude repositories</Label>
                        <RepositoryPatternList repositoryPatterns={job.excludedRepoPatterns} setRepositoryPatterns={updater => {
                            const updatedJob:OwnSignalConfig = {...job, excludedRepoPatterns: updater(job.excludedRepoPatterns)} as OwnSignalConfig
                            onUpdateJob(index, updatedJob)}
                        }/>
                    </div>
                </li>
            ))}
        </Container>
    </div>
)}
