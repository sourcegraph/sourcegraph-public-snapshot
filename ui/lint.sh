#!/bin/bash
set -e

files=$(find ./web_modules -name '*.ts' -and -not -name '*.d.ts')
if [ ${files[@]} ]; then
	echo "Please only use .tsx or .d.ts files:"
	echo $files
	exit 1
fi

if grep -qr --include="*.tsx" propTypes web_modules/; then
	echo "Please use type parameters instead of propTypes:"
	grep -rn --include="*.tsx" propTypes web_modules/
	exit 1
fi

if grep -qr --include="*.tsx" 'className={`' web_modules/; then
	echo "Please use the classNames helper instead of template strings:"
	grep -rn --include="*.tsx" 'className={`' web_modules/
	exit 1
fi

if grep -qr --include="*.tsx" ' from "\.' web_modules/; then
	echo "Please use absolute import paths:"
	grep -rn --include="*.tsx" ' from "\.' web_modules/
	exit 1
fi

# check for TypeScript errors before tslint
#
# We're importing vscode, which is not ready for non-nullable types, but we
# want to do strict null checking in out codebase. Workaround this by running
# the type checker using a config file that does not do strict null checks
# first. Then to another type check with strict null checks turned on but skip
# *.d.ts files, which are the source of the null issues in the vscode project.
./node_modules/.bin/tsc -p tsconfig.lint.json
./node_modules/.bin/tsc --noEmit -p tsconfig.lint.json --skipLibCheck --strictNullChecks
find ./web_modules -name '*.ts' -or -name '*.tsx' | xargs ./node_modules/.bin/tslint
