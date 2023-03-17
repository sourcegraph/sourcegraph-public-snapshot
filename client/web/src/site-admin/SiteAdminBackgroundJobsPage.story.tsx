import {MockedResponse} from '@apollo/client/testing'
import {DecoratorFn, Meta, Story} from '@storybook/react'
import {WildcardMockLink} from 'wildcard-mock-link'

import {getDocumentNode} from '@sourcegraph/http-client'
import {NOOP_TELEMETRY_SERVICE} from '@sourcegraph/shared/src/telemetry/telemetryService'
import {MockedTestProvider} from '@sourcegraph/shared/src/testing/apollo'

import {WebStory} from '../components/WebStory'
import {BACKGROUND_JOBS} from './backend'
import {SiteAdminBackgroundJobsPage} from './SiteAdminBackgroundJobsPage'

const decorator: DecoratorFn = Story => <Story/>

const config: Meta = {
    title: 'web/src/site-admin/SiteAdminBackgroundJobsPage',
    decorators: [decorator],
}

export default config

export const BackgroundJobsPage: Story = () => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(BACKGROUND_JOBS),
                variables: {recentRunCount: 5},
            },
            result: {
                'data': {
                    'backgroundJobs': {
                        'nodes': [{
                            'name': 'auth-permission-sync-job-cleaner', 'routines': [{
                                'name': 'auth.permission_sync_job_cleaner',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Cleans up completed or failed permissions sync jobs',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:22:32Z', 'hostName': 'worker', 'durationMs': 927, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:31Z', 'hostName': 'worker', 'durationMs': 831, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:30Z', 'hostName': 'worker', 'durationMs': 857, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:29Z', 'hostName': 'worker', 'durationMs': 822, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:28Z', 'hostName': 'worker', 'durationMs': 840, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:41Z', 'runCount': 846, 'errorCount': 0, 'minDurationMs': 702, 'avgDurationMs': 923, 'maxDurationMs': 1954, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'auth-permission-sync-job-scheduler', 'routines': [{
                                'name': 'auth.permission_sync_job_scheduler',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Schedule permission sync jobs for users and repositories.',
                                'intervalMs': 10000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:17Z', 'hostName': 'worker', 'durationMs': 1302, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:06Z', 'hostName': 'worker', 'durationMs': 1127, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:55Z', 'hostName': 'worker', 'durationMs': 1245, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:44Z', 'hostName': 'worker', 'durationMs': 1252, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:32Z', 'hostName': 'worker', 'durationMs': 1511, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:07Z', 'runCount': 4666, 'errorCount': 0, 'minDurationMs': 544, 'avgDurationMs': 901, 'maxDurationMs': 6751, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {'name': 'batches-bulk-processor', 'routines': [{'name': 'batches_bulk_processor', 'type': 'DB_BACKED', 'description': 'executes the bulk operations in the background', 'intervalMs': 5000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'batches-janitor',
                            'routines': [{'name': 'batchchanges.cache-cleaner', 'type': 'PERIODIC', 'description': 'cleaning up LRU batch spec execution cache entries', 'intervalMs': 3600000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': '2023-03-17T12:44:23Z', 'runCount': 3, 'errorCount': 0, 'minDurationMs': 3, 'avgDurationMs': 24, 'maxDurationMs': 66, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {
                                'name': 'batchchanges.detached-cleaner',
                                'type': 'PERIODIC',
                                'description': 'cleaning detached changeset entries',
                                'intervalMs': 86400000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [],
                                'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {'name': 'batchchanges.spec-expirer', 'type': 'PERIODIC', 'description': 'expire batch changes specs', 'intervalMs': 3600000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {
                                'name': 'executors.autoscaler-metrics',
                                'type': 'PERIODIC',
                                'description': 'emits metrics to GCP/AWS for auto-scaling',
                                'intervalMs': 5000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 35, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:19Z', 'hostName': 'worker', 'durationMs': 27, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:14Z', 'hostName': 'worker', 'durationMs': 44, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:08Z', 'hostName': 'worker', 'durationMs': 43, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:03Z', 'hostName': 'worker', 'durationMs': 29, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 10228, 'errorCount': 0, 'minDurationMs': 17, 'avgDurationMs': 112, 'maxDurationMs': 7889, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }],
                            '__typename': 'BackgroundJob'
                        }, {'name': 'batches-reconciler', 'routines': [{'name': 'batches_reconciler_worker', 'type': 'DB_BACKED', 'description': 'changeset reconciler that publishes, modifies and closes changesets on the code host', 'intervalMs': 5000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'batches-scheduler', 'routines': [{
                                'name': 'batches-scheduler',
                                'type': 'CUSTOM',
                                'description': 'Scheduler for batch changes',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:22:56Z', 'hostName': 'worker', 'durationMs': 3, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:56Z', 'hostName': 'worker', 'durationMs': 3, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:56Z', 'hostName': 'worker', 'durationMs': 3, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:56Z', 'hostName': 'worker', 'durationMs': 5, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:56Z', 'hostName': 'worker', 'durationMs': 4, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:08Z', 'runCount': 865, 'errorCount': 0, 'minDurationMs': 1, 'avgDurationMs': 12, 'maxDurationMs': 127, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {'name': 'batches-workspace-resolver', 'routines': [{'name': 'batch_changes_batch_spec_resolution_worker', 'type': 'DB_BACKED', 'description': 'runs the workspace resolutions for batch specs, for batch changes running server-side', 'intervalMs': 1000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'bitbucket-project-permissions',
                            'routines': [{'name': 'explicit_permissions_bitbucket_projects_jobs_worker', 'type': 'DB_BACKED', 'description': 'syncs Bitbucket Projects via Explicit Permissions API', 'intervalMs': 1000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}],
                            '__typename': 'BackgroundJob'
                        }, {
                            'name': 'codeintel-autoindexing-dependency-scheduler', 'routines': [{
                                'name': 'precise_code_intel_dependency_indexing_scheduler_worker',
                                'type': 'DB_BACKED',
                                'description': 'queues code-intel auto-indexing jobs for dependency packages',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:28Z', 'hostName': 'worker', 'durationMs': 1082, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 64, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 17275, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:08Z', 'hostName': 'worker', 'durationMs': 17434, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:50Z', 'hostName': 'worker', 'durationMs': 14952, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:04Z', 'runCount': 14754, 'errorCount': 101, 'minDurationMs': 2, 'avgDurationMs': 4649, 'maxDurationMs': 219715, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'precise_code_intel_dependency_sync_scheduler_worker',
                                'type': 'DB_BACKED',
                                'description': 'reads dependency package references from code-intel uploads to be synced to the instance',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:25Z', 'hostName': 'worker', 'durationMs': 338, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:42Z', 'hostName': 'worker', 'durationMs': 6119, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:42Z', 'hostName': 'worker', 'durationMs': 76, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:25Z', 'hostName': 'worker', 'durationMs': 17, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:20Z', 'hostName': 'worker', 'durationMs': 324, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:39Z', 'runCount': 6066, 'errorCount': 0, 'minDurationMs': 5, 'avgDurationMs': 911, 'maxDurationMs': 22530, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'codeintel-autoindexing-janitor', 'routines': [{'name': 'codeintel.autoindexing-janitor', 'type': 'PERIODIC', 'description': 'cleanup autoindexing jobs for unknown repos, commits etc', 'intervalMs': 60000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {
                                'name': 'codeintel.autoindexing.janitor.expired',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes old index records',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:22:33Z', 'hostName': 'worker', 'durationMs': 55898, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:37Z', 'hostName': 'worker', 'durationMs': 60199, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:36Z', 'hostName': 'worker', 'durationMs': 53596, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:16:43Z', 'hostName': 'worker', 'durationMs': 48271, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:14:55Z', 'hostName': 'worker', 'durationMs': 45102, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:01:39Z', 'runCount': 451, 'errorCount': 0, 'minDurationMs': 38702, 'avgDurationMs': 54546, 'maxDurationMs': 106418, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.autoindexing.janitor.unknown-commit',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes index records associated with an unknown commit.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:22:58Z', 'hostName': 'worker', 'durationMs': 55252, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:03Z', 'hostName': 'worker', 'durationMs': 77385, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:45Z', 'hostName': 'worker', 'durationMs': 126072, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:15:39Z', 'hostName': 'worker', 'durationMs': 84902, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:13:15Z', 'hostName': 'worker', 'durationMs': 90892, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:31Z', 'runCount': 339, 'errorCount': 4, 'minDurationMs': 26640, 'avgDurationMs': 92978, 'maxDurationMs': 254775, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.autoindexing.janitor.unknown-repository',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes index records associated with an unknown repository.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:27Z', 'hostName': 'worker', 'durationMs': 9721, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:17Z', 'hostName': 'worker', 'durationMs': 9589, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:08Z', 'hostName': 'worker', 'durationMs': 8279, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:59Z', 'hostName': 'worker', 'durationMs': 10128, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:49Z', 'hostName': 'worker', 'durationMs': 10065, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:01:02Z', 'runCount': 740, 'errorCount': 0, 'minDurationMs': 8154, 'avgDurationMs': 9742, 'maxDurationMs': 17344, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'codeintel-autoindexing-scheduler', 'routines': [{
                                'name': 'codeintel.autoindexing-background-scheduler', 'type': 'PERIODIC_WITH_METRICS', 'description': 'schedule autoindexing jobs in the background using defined or inferred configurations', 'intervalMs': 10000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [{
                                    'at': '2023-03-17T14:23:02Z',
                                    'hostName': 'worker',
                                    'durationMs': 64452,
                                    'errorMessage': '14 errors occurred:\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files',
                                    '__typename': 'BackgroundRoutineRecentRun'
                                }, {
                                    'at': '2023-03-17T14:21:47Z',
                                    'hostName': 'worker',
                                    'durationMs': 73047,
                                    'errorMessage': '10 errors occurred:\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* policies.CommitsDescribedByPolicy: exit status 128',
                                    '__typename': 'BackgroundRoutineRecentRun'
                                }, {
                                    'at': '2023-03-17T14:20:24Z',
                                    'hostName': 'worker',
                                    'durationMs': 63473,
                                    'errorMessage': '11 errors occurred:\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files',
                                    '__typename': 'BackgroundRoutineRecentRun'
                                }, {
                                    'at': '2023-03-17T14:19:11Z',
                                    'hostName': 'worker',
                                    'durationMs': 77726,
                                    'errorMessage': '14 errors occurred:\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files',
                                    '__typename': 'BackgroundRoutineRecentRun'
                                }, {
                                    'at': '2023-03-17T14:17:43Z',
                                    'hostName': 'worker',
                                    'durationMs': 68861,
                                    'errorMessage': '21 errors occurred:\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* policies.CommitsDescribedByPolicy: exit status 128\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files\n\t* indexEnqueuer.QueueIndexes: inference limit: requested content for more than 100 (100) files',
                                    '__typename': 'BackgroundRoutineRecentRun'
                                }], 'stats': {'since': '2023-03-17T00:00:32Z', 'runCount': 747, 'errorCount': 741, 'minDurationMs': 44964, 'avgDurationMs': 59052, 'maxDurationMs': 299941, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.autoindexing-ondemand-scheduler',
                                'type': 'PERIODIC',
                                'description': 'schedule autoindexing jobs for explicitly requested repo+revhash combinations',
                                'intervalMs': 30000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:27Z', 'hostName': 'worker', 'durationMs': 4, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:57Z', 'hostName': 'worker', 'durationMs': 4, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:27Z', 'hostName': 'worker', 'durationMs': 5, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:57Z', 'hostName': 'worker', 'durationMs': 4, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:27Z', 'hostName': 'worker', 'durationMs': 7, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:13Z', 'runCount': 1721, 'errorCount': 0, 'minDurationMs': 2, 'avgDurationMs': 23, 'maxDurationMs': 187, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {'name': 'codeintel-autoindexing-summary-builder', 'routines': [{'name': 'codeintel.autoindexing-summary-builder', 'type': 'PERIODIC', 'description': 'build an auto-indexing summary over repositories with high search activity', 'intervalMs': 1800000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [{'at': '2023-03-17T14:20:27Z', 'hostName': 'worker', 'durationMs': 167185, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}], 'stats': {'since': '2023-03-17T12:14:15Z', 'runCount': 5, 'errorCount': 0, 'minDurationMs': 105662, 'avgDurationMs': 142983, 'maxDurationMs': 167185, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'codeintel-commitgraph-updater', 'routines': [{
                                'name': 'codeintel.commitgraph-updater',
                                'type': 'PERIODIC',
                                'description': 'updates the visibility commit graph for dirty repos',
                                'intervalMs': 10000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:22Z', 'hostName': 'worker', 'durationMs': 293, 'errorMessage': 'gitserver.CommitGraph: reading exec output: repository does not exist: github.com/team-aliens/front-design-system', '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:12Z', 'hostName': 'worker', 'durationMs': 292, 'errorMessage': 'gitserver.CommitGraph: reading exec output: repository does not exist: github.com/team-aliens/front-design-system', '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:01Z', 'hostName': 'worker', 'durationMs': 465, 'errorMessage': 'gitserver.CommitGraph: reading exec output: repository does not exist: github.com/team-aliens/front-design-system', '__typename': 'BackgroundRoutineRecentRun'}, {
                                    'at': '2023-03-17T14:22:51Z',
                                    'hostName': 'worker',
                                    'durationMs': 291,
                                    'errorMessage': 'gitserver.CommitGraph: reading exec output: repository does not exist: github.com/team-aliens/front-design-system',
                                    '__typename': 'BackgroundRoutineRecentRun'
                                }, {'at': '2023-03-17T14:22:41Z', 'hostName': 'worker', 'durationMs': 3759, 'errorMessage': 'gitserver.CommitGraph: reading exec output: repository does not exist: github.com/team-aliens/front-design-system', '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:09Z', 'runCount': 3852, 'errorCount': 3852, 'minDurationMs': 57, 'avgDurationMs': 3358, 'maxDurationMs': 208045, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {'name': 'codeintel-crates-syncer', 'routines': [{'name': 'codeintel.crates-syncer', 'type': 'PERIODIC', 'description': 'syncs the crates list from the index to dependency repos table', 'intervalMs': 7200000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [{'at': '2023-03-17T13:44:51Z', 'hostName': 'worker', 'durationMs': 3364311, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}], 'stats': {'since': '2023-03-17T13:44:51Z', 'runCount': 1, 'errorCount': 0, 'minDurationMs': 3364311, 'avgDurationMs': 3364311, 'maxDurationMs': 3364311, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'codeintel-metrics-reporter', 'routines': [{
                                'name': 'executors.autoscaler-metrics',
                                'type': 'PERIODIC',
                                'description': 'emits metrics to GCP/AWS for auto-scaling',
                                'intervalMs': 5000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 37, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:18Z', 'hostName': 'worker', 'durationMs': 35, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:13Z', 'hostName': 'worker', 'durationMs': 65, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:08Z', 'hostName': 'worker', 'durationMs': 33, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:03Z', 'hostName': 'worker', 'durationMs': 51, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 9939, 'errorCount': 0, 'minDurationMs': 23, 'avgDurationMs': 129, 'maxDurationMs': 15703, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'codeintel-package-filter-applicator', 'routines': [{
                                'name': 'codeintel.package-filter-applicator',
                                'type': 'PERIODIC',
                                'description': 'applies package repo filters to all package repo references to precompute their blocked status',
                                'intervalMs': 5000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 3, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:19Z', 'hostName': 'worker', 'durationMs': 3, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:14Z', 'hostName': 'worker', 'durationMs': 2, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:09Z', 'hostName': 'worker', 'durationMs': 6, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:04Z', 'hostName': 'worker', 'durationMs': 42, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:02Z', 'runCount': 10106, 'errorCount': 0, 'minDurationMs': 1, 'avgDurationMs': 595, 'maxDurationMs': 985, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'codeintel-policies-repository-matcher', 'routines': [{
                                'name': 'codeintel.policies-matcher',
                                'type': 'PERIODIC',
                                'description': 'match repositories to autoindexing+retention policies',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:22:05Z', 'hostName': 'worker', 'durationMs': 28410, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:36Z', 'hostName': 'worker', 'durationMs': 26884, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:09Z', 'hostName': 'worker', 'durationMs': 26752, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:17:42Z', 'hostName': 'worker', 'durationMs': 26262, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:16:16Z', 'hostName': 'worker', 'durationMs': 25682, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:06Z', 'runCount': 593, 'errorCount': 0, 'minDurationMs': 23288, 'avgDurationMs': 27097, 'maxDurationMs': 46196, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'codeintel-ranking-file-reference-counter', 'routines': [{
                                'name': 'codeintel.ranking.file-reference-count-mapper',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Joins ranking definition and references together to create document path count records.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:27Z', 'hostName': 'worker', 'durationMs': 10611, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:15Z', 'hostName': 'worker', 'durationMs': 4177, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:10Z', 'hostName': 'worker', 'durationMs': 9190, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:00Z', 'hostName': 'worker', 'durationMs': 1554, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:57Z', 'hostName': 'worker', 'durationMs': 2297, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T13:44:51Z', 'runCount': 333, 'errorCount': 0, 'minDurationMs': 112, 'avgDurationMs': 124129, 'maxDurationMs': 39355274, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.ranking.file-reference-count-reducer',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Aggregates records from `codeintel_ranking_path_counts_inputs` into `codeintel_path_ranks`.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:28Z', 'hostName': 'worker', 'durationMs': 963, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 447, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:25Z', 'hostName': 'worker', 'durationMs': 447, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:23Z', 'hostName': 'worker', 'durationMs': 450, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:22Z', 'hostName': 'worker', 'durationMs': 507, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 47405, 'errorCount': 0, 'minDurationMs': 1, 'avgDurationMs': 32, 'maxDurationMs': 4655, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.ranking.file-reference-count-seed-mapper',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Adds initial zero counts to files that may not contain any known references.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 3631, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:19Z', 'hostName': 'worker', 'durationMs': 3711, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:15Z', 'hostName': 'worker', 'durationMs': 3632, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:10Z', 'hostName': 'worker', 'durationMs': 3585, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:06Z', 'hostName': 'worker', 'durationMs': 3576, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 8839, 'errorCount': 0, 'minDurationMs': 10, 'avgDurationMs': 5549, 'maxDurationMs': 262930, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.ranking.rank-counts-janitor',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes old path count input records.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:27Z', 'hostName': 'worker', 'durationMs': 742, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:25Z', 'hostName': 'worker', 'durationMs': 749, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:23Z', 'hostName': 'worker', 'durationMs': 749, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:22Z', 'hostName': 'worker', 'durationMs': 759, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:20Z', 'hostName': 'worker', 'durationMs': 716, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 31248, 'errorCount': 0, 'minDurationMs': 416, 'avgDurationMs': 594, 'maxDurationMs': 204957, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.ranking.rank-janitor',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes stale ranking data.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:28Z', 'hostName': 'worker', 'durationMs': 5, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:27Z', 'hostName': 'worker', 'durationMs': 4, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 5, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:25Z', 'hostName': 'worker', 'durationMs': 5, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 8, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 51138, 'errorCount': 0, 'minDurationMs': 3, 'avgDurationMs': 14, 'maxDurationMs': 481, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.ranking.symbol-definitions-janitor',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes stale data from the ranking definitions table.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:28Z', 'hostName': 'worker', 'durationMs': 819, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 849, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 807, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:23Z', 'hostName': 'worker', 'durationMs': 808, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:21Z', 'hostName': 'worker', 'durationMs': 855, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:01Z', 'runCount': 29582, 'errorCount': 0, 'minDurationMs': 55, 'avgDurationMs': 292, 'maxDurationMs': 12507, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.ranking.symbol-exporter',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Exports SCIP data to ranking definitions and reference tables.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:12Z', 'hostName': 'worker', 'durationMs': 4513, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:07Z', 'hostName': 'worker', 'durationMs': 4367, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:01Z', 'hostName': 'worker', 'durationMs': 4571, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:56Z', 'hostName': 'worker', 'durationMs': 16071, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:39Z', 'hostName': 'worker', 'durationMs': 4977, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:01Z', 'runCount': 2905, 'errorCount': 0, 'minDurationMs': 707, 'avgDurationMs': 16779, 'maxDurationMs': 360926, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.ranking.symbol-initial-paths-janitor',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes stale data from the ranking initial paths table.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:27Z', 'hostName': 'worker', 'durationMs': 826, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:25Z', 'hostName': 'worker', 'durationMs': 822, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:23Z', 'hostName': 'worker', 'durationMs': 828, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:21Z', 'hostName': 'worker', 'durationMs': 829, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:19Z', 'hostName': 'worker', 'durationMs': 906, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:01Z', 'runCount': 23886, 'errorCount': 0, 'minDurationMs': 657, 'avgDurationMs': 829, 'maxDurationMs': 11400, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {'name': 'codeintel.ranking.symbol-janitor', 'type': 'PERIODIC_WITH_METRICS', 'description': 'Removes stale data from ranking definitions and reference tables.', 'intervalMs': 1000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {
                                'name': 'codeintel.ranking.symbol-references-janitor',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes stale data from the ranking references table.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:28Z', 'hostName': 'worker', 'durationMs': 823, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 805, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 847, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:22Z', 'hostName': 'worker', 'durationMs': 889, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:21Z', 'hostName': 'worker', 'durationMs': 808, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 20731, 'errorCount': 0, 'minDurationMs': 500, 'avgDurationMs': 1010, 'maxDurationMs': 147053, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {'name': 'ranking.file-reference-count-mapper', 'type': 'PERIODIC', 'description': 'maps definitions and references data to path_counts_inputs table in store', 'intervalMs': 1000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {
                                'name': 'ranking.file-reference-count-reducer',
                                'type': 'PERIODIC',
                                'description': 'reduces path_counts_inputs into a count of paths per repository and stores it in path_ranks table in store.',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [],
                                'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {'name': 'ranking.symbol-exporter', 'type': 'PERIODIC', 'description': 'exports SCIP data to ranking definitions and reference tables', 'intervalMs': 1000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'codeintel-upload-backfiller', 'routines': [{
                                'name': 'codeintel.committed-at-backfiller',
                                'type': 'PERIODIC',
                                'description': 'backfills the committed_at column for code-intel uploads',
                                'intervalMs': 10000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:18Z', 'hostName': 'worker', 'durationMs': 1756, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:06Z', 'hostName': 'worker', 'durationMs': 1889, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:54Z', 'hostName': 'worker', 'durationMs': 1935, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:43Z', 'hostName': 'worker', 'durationMs': 1867, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:31Z', 'hostName': 'worker', 'durationMs': 1645, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:07Z', 'runCount': 4247, 'errorCount': 0, 'minDurationMs': 1485, 'avgDurationMs': 1937, 'maxDurationMs': 20302, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {'name': 'codeintel-upload-expirer', 'routines': [{'name': 'codeintel.upload-expirer', 'type': 'PERIODIC', 'description': 'marks uploads as expired based on retention policies', 'intervalMs': 1000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [{'at': '2023-03-17T13:44:51Z', 'hostName': 'worker', 'durationMs': 4359648, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T12:32:10Z', 'hostName': 'worker', 'durationMs': 24437984, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'codeintel-upload-janitor',
                            'routines': [{'name': 'codeintel.reconciler', 'type': 'PERIODIC', 'description': 'reconciles code-intel data drift', 'intervalMs': 60000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {'name': 'codeintel.upload-janitor', 'type': 'PERIODIC', 'description': 'cleans up various code intel upload and metadata', 'intervalMs': 60000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {
                                'name': 'codeintel.uploads.expirer.unreferenced',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Soft-deletes unreferenced upload records that are not protected by any data retention policy.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:13Z', 'hostName': 'worker', 'durationMs': 17429, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:55Z', 'hostName': 'worker', 'durationMs': 19475, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:36Z', 'hostName': 'worker', 'durationMs': 14280, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:21Z', 'hostName': 'worker', 'durationMs': 14706, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:07Z', 'hostName': 'worker', 'durationMs': 14184, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:26Z', 'runCount': 667, 'errorCount': 0, 'minDurationMs': 10712, 'avgDurationMs': 17576, 'maxDurationMs': 43288, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.expirer.unreferenced-graph',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Soft-deletes a tree of externally unreferenced upload records that are not protected by any data retention policy.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:19Z', 'hostName': 'worker', 'durationMs': 205469, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:53Z', 'hostName': 'worker', 'durationMs': 214724, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:14:19Z', 'hostName': 'worker', 'durationMs': 212415, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:09:46Z', 'hostName': 'worker', 'durationMs': 212115, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:05:14Z', 'hostName': 'worker', 'durationMs': 179336, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:42Z', 'runCount': 198, 'errorCount': 0, 'minDurationMs': 170086, 'avgDurationMs': 201339, 'maxDurationMs': 247886, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.hard-deleter',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Deleted data associated with soft-deleted upload records.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:15Z', 'hostName': 'worker', 'durationMs': 4709, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:10Z', 'hostName': 'worker', 'durationMs': 173, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:10Z', 'hostName': 'worker', 'durationMs': 127, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:10Z', 'hostName': 'worker', 'durationMs': 206, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:09Z', 'hostName': 'worker', 'durationMs': 281, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:48Z', 'runCount': 786, 'errorCount': 0, 'minDurationMs': 107, 'avgDurationMs': 5695, 'maxDurationMs': 900502, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.janitor.abandoned',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes upload records that did did not receive a full payload from the user.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:22:56Z', 'hostName': 'worker', 'durationMs': 20, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:56Z', 'hostName': 'worker', 'durationMs': 2, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:56Z', 'hostName': 'worker', 'durationMs': 2, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:56Z', 'hostName': 'worker', 'durationMs': 2, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:56Z', 'hostName': 'worker', 'durationMs': 2, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:10Z', 'runCount': 861, 'errorCount': 0, 'minDurationMs': 1, 'avgDurationMs': 8, 'maxDurationMs': 129, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.janitor.audit-logs',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Deletes sufficiently old upload audit log records.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:22:30Z', 'hostName': 'worker', 'durationMs': 3232, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:27Z', 'hostName': 'worker', 'durationMs': 3226, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:23Z', 'hostName': 'worker', 'durationMs': 1796, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:21Z', 'hostName': 'worker', 'durationMs': 1516, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:20Z', 'hostName': 'worker', 'durationMs': 1927, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:45Z', 'runCount': 828, 'errorCount': 0, 'minDurationMs': 1481, 'avgDurationMs': 2318, 'maxDurationMs': 14640, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.janitor.scip-documents',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Deletes SCIP document payloads that are not referenced by any index.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:11Z', 'hostName': 'worker', 'durationMs': 408, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:10Z', 'hostName': 'worker', 'durationMs': 334, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:10Z', 'hostName': 'worker', 'durationMs': 364, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:10Z', 'hostName': 'worker', 'durationMs': 382, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:09Z', 'hostName': 'worker', 'durationMs': 397, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:13Z', 'runCount': 857, 'errorCount': 0, 'minDurationMs': 149, 'avgDurationMs': 310, 'maxDurationMs': 1339, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.janitor.unknown-commit',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes upload records associated with an unknown commit.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 16879, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:09Z', 'hostName': 'worker', 'durationMs': 18141, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:51Z', 'hostName': 'worker', 'durationMs': 16508, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:35Z', 'hostName': 'worker', 'durationMs': 23366, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:11Z', 'hostName': 'worker', 'durationMs': 17380, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:43Z', 'runCount': 662, 'errorCount': 6, 'minDurationMs': 13490, 'avgDurationMs': 18156, 'maxDurationMs': 117364, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.janitor.unknown-repository',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes upload records associated with an unknown repository.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:01Z', 'hostName': 'worker', 'durationMs': 8247, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:53Z', 'hostName': 'worker', 'durationMs': 8419, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:44Z', 'hostName': 'worker', 'durationMs': 8404, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:36Z', 'hostName': 'worker', 'durationMs': 9689, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:26Z', 'hostName': 'worker', 'durationMs': 8282, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:01:05Z', 'runCount': 680, 'errorCount': 0, 'minDurationMs': 8083, 'avgDurationMs': 15867, 'maxDurationMs': 35619, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.reconciler.scip-data',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Removes SCIP data records for which there is no known associated metadata in the frontend schema.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:05Z', 'hostName': 'worker', 'durationMs': 3176, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:02Z', 'hostName': 'worker', 'durationMs': 2997, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:59Z', 'hostName': 'worker', 'durationMs': 2991, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:56Z', 'hostName': 'worker', 'durationMs': 4319, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:18:51Z', 'hostName': 'worker', 'durationMs': 3425, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:48Z', 'runCount': 819, 'errorCount': 0, 'minDurationMs': 2563, 'avgDurationMs': 3055, 'maxDurationMs': 8994, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'codeintel.uploads.reconciler.scip-metadata',
                                'type': 'PERIODIC_WITH_METRICS',
                                'description': 'Counts SCIP metadata records for which there is no data in the codeintel-db schema.',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:17Z', 'hostName': 'worker', 'durationMs': 2346, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:14Z', 'hostName': 'worker', 'durationMs': 2369, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:12Z', 'hostName': 'worker', 'durationMs': 2062, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:10Z', 'hostName': 'worker', 'durationMs': 2034, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:08Z', 'hostName': 'worker', 'durationMs': 1863, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:08Z', 'runCount': 792, 'errorCount': 0, 'minDurationMs': 1799, 'avgDurationMs': 5236, 'maxDurationMs': 22859, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }],
                            '__typename': 'BackgroundJob'
                        }, {'name': 'codeintel-uploadstore-expirer', 'routines': [{'name': 'codeintel.upload-store-expirer', 'type': 'PERIODIC', 'description': 'expires entries in the code intel upload store', 'intervalMs': 3600000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'codemonitors-job', 'routines': [{'name': 'code_monitors.trigger_jobs_log_deleter', 'type': 'PERIODIC', 'description': 'deletes code job logs from code monitor triggers', 'intervalMs': 3600000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': '2023-03-17T12:49:35Z', 'runCount': 2, 'errorCount': 0, 'minDurationMs': 29954, 'avgDurationMs': 30323, 'maxDurationMs': 30692, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {
                                'name': 'code_monitors.trigger_query_enqueuer',
                                'type': 'PERIODIC',
                                'description': 'enqueues code monitor trigger query jobs',
                                'intervalMs': 60000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:00Z', 'hostName': 'worker', 'durationMs': 57, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:00Z', 'hostName': 'worker', 'durationMs': 46, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:21:00Z', 'hostName': 'worker', 'durationMs': 26, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:20:00Z', 'hostName': 'worker', 'durationMs': 39, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:19:00Z', 'hostName': 'worker', 'durationMs': 29, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:19Z', 'runCount': 859, 'errorCount': 0, 'minDurationMs': 24, 'avgDurationMs': 57, 'maxDurationMs': 1300, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'code_monitors_action_jobs_worker',
                                'type': 'DB_BACKED',
                                'description': 'runs actions for code monitors',
                                'intervalMs': 5000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:21:51Z', 'hostName': 'worker', 'durationMs': 435, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:09:50Z', 'hostName': 'worker', 'durationMs': 375, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:09:24Z', 'hostName': 'worker', 'durationMs': 503, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:08:49Z', 'hostName': 'worker', 'durationMs': 419, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:07:28Z', 'hostName': 'worker', 'durationMs': 438, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T12:12:57Z', 'runCount': 97, 'errorCount': 48, 'minDurationMs': 9, 'avgDurationMs': 4044, 'maxDurationMs': 32285, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'code_monitors_trigger_jobs_worker',
                                'type': 'DB_BACKED',
                                'description': 'runs trigger queries for code monitors',
                                'intervalMs': 5000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 13401, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:16Z', 'hostName': 'worker', 'durationMs': 84, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:16Z', 'hostName': 'worker', 'durationMs': 326, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:16Z', 'hostName': 'worker', 'durationMs': 130, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:15Z', 'hostName': 'worker', 'durationMs': 63, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 57827, 'errorCount': 9836, 'minDurationMs': 9, 'avgDurationMs': 7108, 'maxDurationMs': 20593, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {'name': 'context-detection-embedding-job', 'routines': [{'name': 'context_detection_embedding_job_worker', 'type': 'DB_BACKED', 'description': '', 'intervalMs': 60000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'executors-janitor', 'routines': [{
                                'name': 'executor.heartbeat-janitor',
                                'type': 'PERIODIC',
                                'description': 'clean up executor heartbeat records for presumed dead executors',
                                'intervalMs': 1800000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:14:56Z', 'hostName': 'worker', 'durationMs': 52, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T13:44:56Z', 'hostName': 'worker', 'durationMs': 51, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T13:44:23Z', 'hostName': 'worker', 'durationMs': 25, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T13:14:23Z', 'hostName': 'worker', 'durationMs': 51, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T12:44:23Z', 'hostName': 'worker', 'durationMs': 25, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T12:14:23Z', 'runCount': 6, 'errorCount': 0, 'minDurationMs': 25, 'avgDurationMs': 40, 'maxDurationMs': 52, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'outbound-webhook-sender',
                            'routines': [{'name': 'outbound-webhooks.janitor', 'type': 'PERIODIC', 'description': 'cleans up stale outbound webhook jobs', 'intervalMs': 3600000, 'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [{'at': '2023-03-17T13:44:56Z', 'hostName': 'worker', 'durationMs': 62, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T13:44:23Z', 'hostName': 'worker', 'durationMs': 1, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T12:44:23Z', 'hostName': 'worker', 'durationMs': 1, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}, {
                                'name': 'outbound_webhook_job_worker',
                                'type': 'DB_BACKED',
                                'description': '',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': '2023-03-17T13:44:51Z', '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [],
                                'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }],
                            '__typename': 'BackgroundJob'
                        }, {
                            'name': 'record-encrypter', 'routines': [{
                                'name': 'encryption.operation-metrics',
                                'type': 'PERIODIC',
                                'description': 'tracks number of encrypted vs unencrypted records',
                                'intervalMs': 10000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 218, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:14Z', 'hostName': 'worker', 'durationMs': 73, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:04Z', 'hostName': 'worker', 'durationMs': 65, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:53Z', 'hostName': 'worker', 'durationMs': 84, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:22:43Z', 'hostName': 'worker', 'durationMs': 66, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:06Z', 'runCount': 5123, 'errorCount': 0, 'minDurationMs': 49, 'avgDurationMs': 78, 'maxDurationMs': 366, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }, {
                                'name': 'encryption.record-encrypter',
                                'type': 'PERIODIC',
                                'description': 'encrypts/decrypts existing data when a key is provided/removed',
                                'intervalMs': 1000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:23:28Z', 'hostName': 'worker', 'durationMs': 66, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:26Z', 'hostName': 'worker', 'durationMs': 74, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:25Z', 'hostName': 'worker', 'durationMs': 63, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:24Z', 'hostName': 'worker', 'durationMs': 58, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T14:23:23Z', 'hostName': 'worker', 'durationMs': 52, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T00:00:00Z', 'runCount': 48751, 'errorCount': 0, 'minDurationMs': 20, 'avgDurationMs': 60, 'maxDurationMs': 498, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {'name': 'repo-embedding-job', 'routines': [{'name': 'repo_embedding_job_worker', 'type': 'DB_BACKED', 'description': '', 'intervalMs': 60000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [], 'stats': {'since': null, 'runCount': 0, 'errorCount': 0, 'minDurationMs': 0, 'avgDurationMs': 0, 'maxDurationMs': 0, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}, {
                            'name': 'repo-statistics-compactor', 'routines': [{
                                'name': 'repomgmt.statistics-compactor',
                                'type': 'PERIODIC',
                                'description': 'compacts repo statistics',
                                'intervalMs': 1800000,
                                'instances': [{'hostName': 'worker', 'lastStartedAt': null, 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}],
                                'recentRuns': [{'at': '2023-03-17T14:14:56Z', 'hostName': 'worker', 'durationMs': 30, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T13:44:56Z', 'hostName': 'worker', 'durationMs': 69, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T13:44:24Z', 'hostName': 'worker', 'durationMs': 55, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T13:14:24Z', 'hostName': 'worker', 'durationMs': 62, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T12:44:24Z', 'hostName': 'worker', 'durationMs': 44, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}],
                                'stats': {'since': '2023-03-17T12:14:24Z', 'runCount': 6, 'errorCount': 0, 'minDurationMs': 30, 'avgDurationMs': 60, 'maxDurationMs': 98, '__typename': 'BackgroundRoutineStats'},
                                '__typename': 'BackgroundRoutine'
                            }], '__typename': 'BackgroundJob'
                        }, {
                            'name': 'webhook-log-janitor',
                            'routines': [{'name': 'batchchanges.webhook-log-janitor', 'type': 'PERIODIC', 'description': 'cleans up stale webhook logs', 'intervalMs': 3600000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [{'at': '2023-03-17T13:44:56Z', 'hostName': 'worker', 'durationMs': 44, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T13:44:23Z', 'hostName': 'worker', 'durationMs': 1, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T12:44:23Z', 'hostName': 'worker', 'durationMs': 2, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}], 'stats': {'since': '2023-03-17T12:44:23Z', 'runCount': 3, 'errorCount': 0, 'minDurationMs': 1, 'avgDurationMs': 16, 'maxDurationMs': 44, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}],
                            '__typename': 'BackgroundJob'
                        }, {'name': 'zoekt-repos-updater', 'routines': [{'name': 'search.index-status-reconciler', 'type': 'PERIODIC', 'description': 'reconciles indexed status between zoekt and postgres', 'intervalMs': 3600000, 'instances': [{'hostName': 'worker', 'lastStartedAt': '2023-03-17T13:44:56Z', 'lastStoppedAt': null, '__typename': 'BackgroundRoutineInstance'}], 'recentRuns': [{'at': '2023-03-17T13:45:46Z', 'hostName': 'worker', 'durationMs': 50212, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}, {'at': '2023-03-17T12:53:19Z', 'hostName': 'worker', 'durationMs': 50185, 'errorMessage': null, '__typename': 'BackgroundRoutineRecentRun'}], 'stats': {'since': '2023-03-17T12:53:19Z', 'runCount': 2, 'errorCount': 0, 'minDurationMs': 50185, 'avgDurationMs': 50199, 'maxDurationMs': 50212, '__typename': 'BackgroundRoutineStats'}, '__typename': 'BackgroundRoutine'}], '__typename': 'BackgroundJob'}], '__typename': 'BackgroundJobConnection'
                    }
                }
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])
    return (
        <WebStory>
            {() => (
                <MockedTestProvider link={mocks}>
                    <SiteAdminBackgroundJobsPage telemetryService={NOOP_TELEMETRY_SERVICE}/>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

BackgroundJobsPage.storyName = 'Background jobs'

export const BackgroundJobsPageWithError: Story = () => {
    const mockedResponse: MockedResponse[] = [
        {
            request: {
                query: getDocumentNode(BACKGROUND_JOBS),
                variables: {first: null, after: null},
            },
            error: new Error('oops'),
        },
    ]
    return (
        <WebStory>
            {() => (
                <MockedTestProvider mocks={mockedResponse}>
                    <SiteAdminBackgroundJobsPage telemetryService={NOOP_TELEMETRY_SERVICE}/>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

BackgroundJobsPageWithError.storyName = 'Error during background jobs fetch'
