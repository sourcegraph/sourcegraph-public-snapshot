#!/bin/bash -u

LSIF_TSC=${LSIF_TSC:-`which lsif-tsc`}

trap '{ rm -r ./repo; }' EXIT

mkdir -p repo/src

cat << EOF > 'repo/src/index.ts'
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

cat << EOF > 'repo/tsconfig.json'
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

mkdir -p data
${LSIF_TSC} -p repo/tsconfig.json --noContents --out ./data/test.lsif
gzip ./data/*.lsif
