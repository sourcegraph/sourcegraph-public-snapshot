#!/bin/bash -u

mkdir -p "${DIR}/${REPO}/src"
mkdir -p "${DIR}/${REPO}/node_modules"

# Link the math-util node module (unpublished)
ln -s ${DEP} "${DIR}/${REPO}/node_modules/math-util"

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
        "math-util": "^0.1.0"
    },
    "scripts": {
        "build": "tsc"
    }
}
EOF

cat << EOF > "${DIR}/${REPO}/tsconfig.json"
{
    "compilerOptions": {},
    "include": ["src"],
    "exclude": ["node_modules"],
}
EOF
