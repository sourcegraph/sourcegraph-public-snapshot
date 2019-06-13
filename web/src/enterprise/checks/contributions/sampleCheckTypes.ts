import AppleFinderIcon from 'mdi-react/AppleFinderIcon'
import DeleteSweepIcon from 'mdi-react/DeleteSweepIcon'
import DockerIcon from 'mdi-react/DockerIcon'
import LanguageGoIcon from 'mdi-react/LanguageGoIcon'
import LanguagePythonIcon from 'mdi-react/LanguagePythonIcon'
import LanguageTypescriptIcon from 'mdi-react/LanguageTypescriptIcon'
import NpmIcon from 'mdi-react/NpmIcon'
import ReactIcon from 'mdi-react/ReactIcon'
import { CheckTemplate } from '../../../../../shared/src/api/client/services/checkTemplates'

export const CHECK_TYPES: CheckTemplate[] = [
    {
        id: 'check.upgradeNpmDependency',
        title: 'Upgrade npm dependency',
        description: 'Ensure an npm dependency is at a specific version',
        icon: NpmIcon,
        iconColor: '#cb3837',
    },
    {
        id: 'check.npmPackageAllowlist',
        title: 'Require approval for new npm package dependencies',
        description:
            'Define the approved set of npm packages, find dependencies on unapproved packages, and prevent new violations',
        icon: NpmIcon,
        iconColor: '#cb3837',
    },
    {
        id: 'check.typescriptTSConfig',
        title: 'Standardize TypeScript tsconfig.json files',
        description: "Enforce consistency among TypeScript projects' tsconfig.json files",
        icon: LanguageTypescriptIcon,
        iconColor: '#2774c3',
    },
    {
        id: 'check.dockerfileLint',
        title: 'Dockerfile lint',
        description: 'Find and fix common mistakes in Dockerfiles',
        icon: DockerIcon,
        iconColor: '#0db7ed',
    },
    {
        id: 'check.goLint',
        title: 'Go lint',
        description: 'Fix common mistakes in Go code',
        icon: LanguageGoIcon,
    },
    {
        id: 'check.goPackageImportsAllowlist',
        title: 'Require approval for Go package imports',
        description: 'Find imports of unapproved Go packages and prevent new violations from being committed',
        icon: LanguageGoIcon,
    },
    {
        id: 'check.reactLint',
        title: 'React lint',
        description: 'Fix common problems in React code & migrate deprecated React code',
        icon: ReactIcon,
        iconColor: '#00d8ff',
        settings: {
            queries: [
                'file:\\.[tj]sx$ key="hardcoded timeout:10s',
                'file:\\.[tj]sx$ defaultValue=',
                "file:\\.[tj]sx$ \\w+=\\{[`'][^$'`]*[`']\\}",
            ],
        },
    },
    {
        id: 'check.removeDSStore',
        title: 'No macOS .DS_Store files',
        description: 'Deletes and gitignores undesired macOS temp and metadata files',
        icon: AppleFinderIcon,
    },
    {
        id: 'check.removePYCFiles',
        title: 'No *.pyc files',
        description: 'Deletes and gitignores undesired Python temp files',
        icon: LanguagePythonIcon,
    },
    {
        id: 'check.removeYarnNpmTempFiles',
        title: 'No undesired Yarn/npm temporary files',
        description: 'Deletes and gitignores yarn-error.log and npm-debug.log files',
        icon: DeleteSweepIcon,
    },
]
