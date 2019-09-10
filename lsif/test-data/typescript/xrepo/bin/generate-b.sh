#!/bin/bash -u

mkdir -p "${DIR}/${REPO}/src"

cat << EOF > "${DIR}/${REPO}/src/index.ts"
import { add, mul } from 'math-util/src'

export function foobar(a: number, b: number): number {
    return add(mul(a, b), mul(b, a))
}
EOF

cat << EOF > "${DIR}/${REPO}/package.json"
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

cat << EOF > "${DIR}/${REPO}/tsconfig.json"
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

yarn --cwd "${DIR}/${REPO}" > /dev/null
