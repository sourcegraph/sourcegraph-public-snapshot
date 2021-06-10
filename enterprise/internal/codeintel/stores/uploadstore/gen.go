package uploadstore

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/derision-test/go-mockgen/cmd/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore -i s3API -i s3Uploader -o mock_s3_api_test.go -p uploadstore
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore -i gcsAPI -i gcsBucketHandle -i gcsObjectHandle -i gcsComposer -o mock_gcs_api_test.go -p uploadstore
