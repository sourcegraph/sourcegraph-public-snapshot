package maven

type CoursierClient struct {}

func (c *CoursierClient) ListGroupIDs() []string {

}

func (c *CoursierClient) ListArtifactIDs(groupID string) []string {

}

func (c *CoursierClient) ListVersions(groupID, artifactID string) []string {

}

func (c *CoursierClient) FetchVersions(groupID, artifactID string, versions []string) []string {

}

func (c *CoursierClient) Exists(groupID, artifactID string) bool {
	return len(c.ListVersions(groupID, artifactID)) > 0
}
