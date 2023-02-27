package linters

var (
	goEnterpriseImport = runScript("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh")
	tsEnterpriseImport = runScript("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh")
)
