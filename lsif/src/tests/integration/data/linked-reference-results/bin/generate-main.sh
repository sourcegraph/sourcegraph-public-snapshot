#!/usr/bin/env bash -u
#
# This script runs in the parent directory.
# This script is invoked by ./generate.sh.
# Do NOT run directly.

mkdir -p ./repos/main/src

cat << EOF > './repos/main/src/index.ts'
interface I {
    foo(): void;
}

class A implements I {
    foo(): void {}
}

class B implements I {
    foo(): void {}
}

let i: I;
i.foo();

let b: B;
b.foo();
EOF

cat << EOF > './repos/main/tsconfig.json'
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
