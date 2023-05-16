import { execSync } from 'child_process'

interface Scaffold {
    scaffold_name: string
    dependencies: string[]
    command: string
}

interface ScaffoldStatus {
    generatorStatus: boolean
    generatorMessage: string
}

export class ScaffoldGenerator {
    public generateReactApp(rootPath: string, appName: string): ScaffoldStatus {
        let status = false
        let message = ''
        const react: Scaffold = {
            scaffold_name: 'React app generator',
            dependencies: ['node -v', 'npx -v', 'npm -v'],
            command: 'npx create-react-app',
        }

        // check for the dependencies first
        for (const dependency of react.dependencies) {
            try {
                execSync(dependency, { cwd: rootPath })
                status = true
            } catch (error) {
                status = false
                message = error
            }
        }

        // if all the dependency are present then run create command
        if (status) {
            try {
                execSync(`${react.command} ${appName}`, { cwd: rootPath })
                message = 'React app generated successfully'
                status = true
            } catch (error) {
                status = false
                message = error
            }
        }

        return {
            generatorStatus: status,
            generatorMessage: message,
        }
    }
}
