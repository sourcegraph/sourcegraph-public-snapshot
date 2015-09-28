package ssh

import (
	"strings"
	"testing"
)

func TestParseKnownHosts_ok(t *testing.T) {
	data := `
xenon.stanford.edu,171.64.66.201 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAsWfJKexGQfH549CzZHbmaGRJ1307nkCADIJqmnZQpMSWiE1yGxOWevjYMv4nxqefQko8W3ixNTUs0dzFvmxImAqNId6F8RBW3jt7rj6o1+L9VNCx2UtWUtr0CXifAUnef2iPoT3vS50IkArHp71M8fDruH5wbPcbnP76odGfODWJU2qcNHIMbLoUuxULUHSzCzM+kOVCC9nl7P1OJUbsvuw5mjBJbFRbQW1Zctny1lyRlftDGUjYYBR5G18qtn6w0+w9OhCoSAFd1bQq982kfgVIRQokhLC7Eq24cQTKT85zN/m8I9lptkxWGsHcTV9nMG+LKv2pbE3JOPqwR/556Q==
|1|Lr7o99feGO4XWwfc09dxyiY/nMo=|TRs4gnNyZS1i1QbBW5XvGbjr1R8= ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==
 
|1|k4TmSR1fbu8pZ9cubTOMXcQguUc=|SsAqIoaKcq/KzLN7o2pVdToZPzs= ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==
`
	kh, err := ParseKnownHosts(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := kh.Lookup("xenon.stanford.edu"); !ok {
		t.Error("got no xenon.stanford.edu host key, want 1")
	}
	if _, ok := kh.Lookup("171.64.66.201"); !ok {
		t.Error("got no xenon.stanford.edu host key, want 1")
	}
	if _, ok := kh.Lookup("github.com"); !ok {
		t.Error("got no github.com host key, want 1")
	}
	if _, ok := kh.Lookup("doesntexist.com"); ok {
		t.Error("got doesntexist.com host key, want none")
	}
}

func TestParseKnownHosts_invalidFormat(t *testing.T) {
	data := `
bad format
`

	if _, err := ParseKnownHosts(strings.NewReader(data)); err == nil {
		t.Fatal("got err == nil, want non-nil err")
	}
}
