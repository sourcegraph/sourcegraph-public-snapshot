#!/usr/bin/env bash -u
#
# This script runs in the parent directory.
# This script is invoked by ./generate.sh.
# Do NOT run directly.

mkdir -p "./repos/${REPO}/src"

cat << EOF > "./repos/${REPO}/src/index.ts"
import { add } from 'math-util/src'

// Peano-construction of 5
add(1, add(1, add(1, add(1, 1))))
EOF

cat << EOF > "./repos/${REPO}/package.json"
{
    "name": "${REPO}",
    "license": "MIT",
    "version": "0.1.0",
    "dependencies": {
        "math-util": "link:${DEP}"
    },
    "scripts": {
        "build": "tsc"
    }
}
EOF

cat << EOF > "./repos/${REPO}/tsconfig.json"
{
    "compilerOptions": {
        "module": "commonjs",
        "target": "esnext",
        "moduleResolution": "node",
        "typeRoots": []
    },
    "include": ["src/*"],
    "exclude": ["node_modules"]
}
EOF

yarn --cwd "./repos/${REPO}" > /dev/null
