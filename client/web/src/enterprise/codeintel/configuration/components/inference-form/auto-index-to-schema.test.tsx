import AJV from 'ajv'
import addFormats from 'ajv-formats'
import { describe, expect, it } from 'vitest'

import type { AutoIndexJobDescriptionFields } from '../../../../../graphql-operations'
import schema from '../../schema.json'

import { autoIndexJobsToFormData } from './auto-index-to-form-job'
import { formDataToSchema } from './form-data-to-schema'

const ajv = new AJV({ strict: false })
addFormats(ajv)

const mockAutoIndexJobs: AutoIndexJobDescriptionFields[] = [
    {
        comparisonKey: 'key-1',
        indexer: {
            imageName: 'indexer-1@123',
            name: 'indexer-1',
            url: 'https://example.com',
            key: 'key-indexer-1',
        },
        root: 'root-1',
        steps: {
            index: {
                commands: ['command-1', 'command-2'],
                indexerArgs: ['arg-1', 'arg-2'],
                requestedEnvVars: ['ENV_VAR', 'ENV_VAR_2'],
                outfile: 'outfile-1',
            },
            preIndex: [
                {
                    commands: ['pre-command-1', 'pre-command-2'],
                    root: 'pre-root-1',
                    image: 'indexer-1@123',
                },
            ],
        },
    },
    {
        comparisonKey: 'key-2',
        indexer: null,
        root: '',
        steps: {
            index: {
                commands: [],
                indexerArgs: [],
                requestedEnvVars: null,
                outfile: null,
            },
            preIndex: [],
        },
    },
]

describe('autoIndexToAutoIndexSchema', () => {
    it('should build form data as expected', () => {
        const formData = autoIndexJobsToFormData({ jobs: mockAutoIndexJobs })
        expect(formData).toMatchSnapshot()
    })

    it('should build schema data as expected', () => {
        const formData = autoIndexJobsToFormData({ jobs: mockAutoIndexJobs })
        const schemaData = formDataToSchema(formData)
        expect(schemaData).toMatchSnapshot()
    })

    it('should build valid schema data', () => {
        const formData = autoIndexJobsToFormData({ jobs: mockAutoIndexJobs })
        const schemaData = formDataToSchema(formData)

        // Validate form data against JSONSchema
        const isValid = ajv.validate(schema, schemaData)

        expect(isValid).toBe(true)
    })
})
