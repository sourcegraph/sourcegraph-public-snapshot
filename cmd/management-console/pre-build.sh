set -ex

pushd web/
npm install
npm run build
popd
go generate ./assets
