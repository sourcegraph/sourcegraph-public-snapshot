import React, {FC, useState} from 'react';
import {PageTitle} from '../../../components/PageTitle';
import {
    Container,
    PageHeader,
    H3, Text, Label, Button, LoadingSpinner, ErrorAlert
} from '@sourcegraph/wildcard'
import styles from '../../insights/admin-ui/CodeInsightsJobs.module.scss';
import './own-status-page-styles.scss'
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import {RepositoryPatternList} from '../../codeintel/configuration/components/RepositoryPatternList';
import {useQuery} from '@sourcegraph/http-client';
import {
    GetOwnSignalConfigurationsResult,
} from '../../../graphql-operations';
import {GET_OWN_JOB_CONFIGURATIONS} from './query';

interface Job {
    name: string;
    description: string;
    isEnabled: boolean;
    excluded: string[];  // array of strings
}

const data: Job[] = [{
    'name': 'recent-contributors',
    'description': 'Calculates recent contributors one job per repository.',
    'isEnabled': true,
    excluded: []
},
    {
        'name': 'recent-views',
        'description': 'Calculates recent viewers from the events stored inside Sourcegraph.',
        'isEnabled': false,
        excluded: []
    }
]

export const OwnStatusPage: FC = () => {
    const [hasLocalChanges, setHasLocalChanges] = useState<boolean>(false)
    const [localData, setLocalData] = useState<Job[]>([])

    const { remoteData, loading, error } = useQuery<GetOwnSignalConfigurationsResult>(
        GET_OWN_JOB_CONFIGURATIONS, {onCompleted: data => {
                console.log(data)
                const jobs = data.signalConfigurations.map(sc => {
                    return {...sc, excluded: sc.excludedRepoPatterns}
                })
                setLocalData(jobs)
            }}
    )

    function onUpdateJob(index: number, newJob: Job): void {
        setHasLocalChanges(true)
        const newData = localData.map((job, ind) => {
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
                <PageTitle title="Own status page"/>
                <PageHeader
                    headingElement="h2"
                    path={[{text: 'Own status page'}]}
                    description="List of Own inference signal indexers and their status. All repositories are included by default."
                    className="mb-3"
                />
            </div>

            <Button id='saveButton' disabled={!hasLocalChanges} aria-label="Save changes" variant="primary" onClick={() => {
                // do network stuff
                setHasLocalChanges(false)
            }}>Save</Button>
        </span>

        <Container className={styles.root}>
            {loading && <LoadingSpinner/>}
            {error && <ErrorAlert prefix="Error fetching Own signal configurations" error={error} /> }
            {!loading && localData && !error && localData.map((job, index) => (
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
                        <RepositoryPatternList repositoryPatterns={job.excluded} setRepositoryPatterns={updater => {
                            onUpdateJob(index, {...job, excluded: updater(job.excluded)} as Job)}
                        }/>

                    </div>
                </li>
            ))}
        </Container>
    </div>
)}
