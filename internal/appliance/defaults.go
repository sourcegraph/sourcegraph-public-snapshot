package appliance

func newDefaultConfig() Sourcegraph {
	return Sourcegraph{
		Spec: SourcegraphSpec{
			Blobstore: BlobstoreSpec{
				StorageSize: "100Gi",
			},
			StorageClass: StorageClassSpec{
				Name: "sourcegraph",
			},
			Symbols: SymbolsSpec{
				Replicas:    1,
				StorageSize: "12Gi",
			},
		},
	}
}
