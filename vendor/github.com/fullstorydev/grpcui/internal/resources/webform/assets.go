package webform

func Template() []byte {
	return MustAsset("webform-template.html")
}

func Script() []byte {
	return MustAsset("webform.js")
}

func SampleCSS() []byte {
	return MustAsset("webform-sample.css")
}
