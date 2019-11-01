#!/bin/bash -u

mkdir -p "${DIR}/${REPO}/src"

cat << EOF > "${DIR}/${REPO}/src/index.ts"
export function add(a: number, b: number): number {
    return a + b
}

// Peano-construction of 5
add(1, add(1, add(1, add(1, 1))))

// Peano-construction of 10
add(1, add(1, add(1, add(1, add(1, add(1, add(1, add(1, add(1, 1)))))))))
EOF

cat << EOF > "${DIR}/${REPO}/package.json"
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
