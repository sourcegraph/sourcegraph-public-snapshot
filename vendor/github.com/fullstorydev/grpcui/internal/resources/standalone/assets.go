package standalone

const IndexTemplateName = "index-template.html"

func IndexTemplate() []byte {
	return MustAsset(IndexTemplateName)
}
