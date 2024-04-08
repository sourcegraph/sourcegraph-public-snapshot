// Standalone script to generate graphql types, used by bazel
import { codegen } from './vite-graphql-codegen'

codegen().catch(error => {
    console.error(error)
    process.exit(1)
})
