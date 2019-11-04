#!/usr/bin/env bash -u
#
# This script runs in the parent directory.
# This script is invoked by ./generate.sh.
# Do NOT run directly.

mkdir -p repos/main/src

cat << EOF > 'repos/main/src/a.ts'
export function add(a: number, b: number): number {
    return a + b
}
EOF

cat << EOF > 'repos/main/src/b.ts'
import { add } from './a'

add(1, add(2, add(3, 4)))
EOF

cat << EOF > 'repos/main/tsconfig.json'
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
