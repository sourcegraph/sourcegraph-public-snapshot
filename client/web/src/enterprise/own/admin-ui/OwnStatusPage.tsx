import {FC, useState} from 'react';
import {PageTitle} from '../../../components/PageTitle';
import {
    Container,
    PageHeader,
    H3, Text, Label, Button
} from '@sourcegraph/wildcard'
import styles from '../../insights/admin-ui/CodeInsightsJobs.module.scss';
import './own-status-page-styles.scss'
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import {RepositoryPatternList} from '../../codeintel/configuration/components/RepositoryPatternList';

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
    const [localData, setLocalData] = useState<Job[]>(data)
    const [hasLocalChanges, setHasLocalChanges] = useState<boolean>(false)

    function onUpdateJob(index, newJob): void {
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
                    description="List of Own signal indexers and their status"
                    className="mb-3"
                />
            </div>

            <Button id='saveButton' disabled={!hasLocalChanges} variant="primary" onClick={() => {
                // do network stuff
                setHasLocalChanges(false)
            }}>Save</Button>
        </span>

        <Container className={styles.root}>
            {localData.map((job, index) => (
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
                            onUpdateJob(index, {...job, excluded: updater(job.excluded)})}
                        }/>

                    </div>
                </li>
            ))}
        </Container>
    </div>
)}
