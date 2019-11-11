#!/usr/bin/env bash -u
#
# This script runs in the parent directory.
# This script is invoked by ./generate.sh.
# Do NOT run directly.

mkdir -p "./repos/${REPO}/src"

cat << EOF > "./repos/${REPO}/src/index.ts"
export function add(a: number, b: number): number {
    return a + b
}

export function mul(a: number, b: number): number {
    if (b === 0) {
        return 0
    }

    let product = a
    for (let i = 0; i < b; i++) {
        product = add(product, a)
    }

    return product
}
EOF

cat << EOF > "./repos/${REPO}/package.json"
{
    "name": "math-util",
    "license": "MIT",
    "version": "0.1.0",
    "dependencies": {},
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
