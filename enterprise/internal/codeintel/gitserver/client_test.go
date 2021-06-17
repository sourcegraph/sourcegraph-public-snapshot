package gitserver

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParseCommitGraph(t *testing.T) {
	graph := ParseCommitGraph([]string{
		"9ad62c7ec68e377b41a8b8dd846e573b76634172 1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3 683cafd122632142bda6e36563f5719e5b0fa37d",
		"683cafd122632142bda6e36563f5719e5b0fa37d 1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3",
		"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3 02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d",
		"02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d a94fb112d1f2e70f55851c6c569916a9e31caee1",
		"a94fb112d1f2e70f55851c6c569916a9e31caee1 2716762a5213f5fe2576d2a52d1182282704004c",
		"2716762a5213f5fe2576d2a52d1182282704004c 7696eb70f56308b7286cf1fb156af6bbb57495e4",
		"7696eb70f56308b7286cf1fb156af6bbb57495e4 33d2968a49e1131825a0accdd64bdf9901cf144c",
		"33d2968a49e1131825a0accdd64bdf9901cf144c 0b3eb4b7b8da63598d99905dd63e1c9ba92aaf22",
		"0b3eb4b7b8da63598d99905dd63e1c9ba92aaf22 843793335d0f8d70c1d9f1c948ff584a979b5bab",
		"843793335d0f8d70c1d9f1c948ff584a979b5bab 61b1debf13ba14e09e29042727d698073bb87f83 5f93bf0eb32807657bc4dafcc934f817991bd805",
		"5f93bf0eb32807657bc4dafcc934f817991bd805 189f2dd3ec0761e60d054f1dc880cfbf074ede16",
		"189f2dd3ec0761e60d054f1dc880cfbf074ede16 61b1debf13ba14e09e29042727d698073bb87f83",
		"61b1debf13ba14e09e29042727d698073bb87f83 f9fdcbb3871320e0655530b707a3d276d561311a",
		"f9fdcbb3871320e0655530b707a3d276d561311a 6d238ad929833c066db4fb3305e4614c212efb42",
		"6d238ad929833c066db4fb3305e4614c212efb42 c855987d88ef32861b61bfa1118781e63a7b0457",
		"c855987d88ef32861b61bfa1118781e63a7b0457 267fa34a98ba30e53d228500831b6db37883c199",
		"267fa34a98ba30e53d228500831b6db37883c199 7b8acaa20dd8db659a31520d95ae74afe0e43499",
		"7b8acaa20dd8db659a31520d95ae74afe0e43499 e883e8a1bf875f6ed5e095e4355acc54ef38ea2c",
		"e883e8a1bf875f6ed5e095e4355acc54ef38ea2c 8c25906a7dbc904acc96b5435ec42fa5da8b232c",
		"8c25906a7dbc904acc96b5435ec42fa5da8b232c",
	})

	expectedGraph := map[string][]string{
		"9ad62c7ec68e377b41a8b8dd846e573b76634172": {"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3", "683cafd122632142bda6e36563f5719e5b0fa37d"},
		"683cafd122632142bda6e36563f5719e5b0fa37d": {"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3"},
		"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3": {"02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d"},
		"02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d": {"a94fb112d1f2e70f55851c6c569916a9e31caee1"},
		"a94fb112d1f2e70f55851c6c569916a9e31caee1": {"2716762a5213f5fe2576d2a52d1182282704004c"},
		"2716762a5213f5fe2576d2a52d1182282704004c": {"7696eb70f56308b7286cf1fb156af6bbb57495e4"},
		"7696eb70f56308b7286cf1fb156af6bbb57495e4": {"33d2968a49e1131825a0accdd64bdf9901cf144c"},
		"33d2968a49e1131825a0accdd64bdf9901cf144c": {"0b3eb4b7b8da63598d99905dd63e1c9ba92aaf22"},
		"0b3eb4b7b8da63598d99905dd63e1c9ba92aaf22": {"843793335d0f8d70c1d9f1c948ff584a979b5bab"},
		"843793335d0f8d70c1d9f1c948ff584a979b5bab": {"61b1debf13ba14e09e29042727d698073bb87f83", "5f93bf0eb32807657bc4dafcc934f817991bd805"},
		"5f93bf0eb32807657bc4dafcc934f817991bd805": {"189f2dd3ec0761e60d054f1dc880cfbf074ede16"},
		"189f2dd3ec0761e60d054f1dc880cfbf074ede16": {"61b1debf13ba14e09e29042727d698073bb87f83"},
		"61b1debf13ba14e09e29042727d698073bb87f83": {"f9fdcbb3871320e0655530b707a3d276d561311a"},
		"f9fdcbb3871320e0655530b707a3d276d561311a": {"6d238ad929833c066db4fb3305e4614c212efb42"},
		"6d238ad929833c066db4fb3305e4614c212efb42": {"c855987d88ef32861b61bfa1118781e63a7b0457"},
		"c855987d88ef32861b61bfa1118781e63a7b0457": {"267fa34a98ba30e53d228500831b6db37883c199"},
		"267fa34a98ba30e53d228500831b6db37883c199": {"7b8acaa20dd8db659a31520d95ae74afe0e43499"},
		"7b8acaa20dd8db659a31520d95ae74afe0e43499": {"e883e8a1bf875f6ed5e095e4355acc54ef38ea2c"},
		"e883e8a1bf875f6ed5e095e4355acc54ef38ea2c": {"8c25906a7dbc904acc96b5435ec42fa5da8b232c"},
		"8c25906a7dbc904acc96b5435ec42fa5da8b232c": {},
	}
	if diff := cmp.Diff(expectedGraph, graph.Graph()); diff != "" {
		t.Errorf("unexpected commit mapping (-want +got):\n%s", diff)
	}

	expectedOrder := []string{
		"8c25906a7dbc904acc96b5435ec42fa5da8b232c",
		"e883e8a1bf875f6ed5e095e4355acc54ef38ea2c",
		"7b8acaa20dd8db659a31520d95ae74afe0e43499",
		"267fa34a98ba30e53d228500831b6db37883c199",
		"c855987d88ef32861b61bfa1118781e63a7b0457",
		"6d238ad929833c066db4fb3305e4614c212efb42",
		"f9fdcbb3871320e0655530b707a3d276d561311a",
		"61b1debf13ba14e09e29042727d698073bb87f83",
		"189f2dd3ec0761e60d054f1dc880cfbf074ede16",
		"5f93bf0eb32807657bc4dafcc934f817991bd805",
		"843793335d0f8d70c1d9f1c948ff584a979b5bab",
		"0b3eb4b7b8da63598d99905dd63e1c9ba92aaf22",
		"33d2968a49e1131825a0accdd64bdf9901cf144c",
		"7696eb70f56308b7286cf1fb156af6bbb57495e4",
		"2716762a5213f5fe2576d2a52d1182282704004c",
		"a94fb112d1f2e70f55851c6c569916a9e31caee1",
		"02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d",
		"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3",
		"683cafd122632142bda6e36563f5719e5b0fa37d",
		"9ad62c7ec68e377b41a8b8dd846e573b76634172",
	}
	if diff := cmp.Diff(expectedOrder, graph.Order()); diff != "" {
		t.Errorf("unexpected commit order (-want +got):\n%s", diff)
	}
}

func TestParseCommitGraphPartial(t *testing.T) {
	graph := ParseCommitGraph([]string{
		"9ad62c7ec68e377b41a8b8dd846e573b76634172 1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3 683cafd122632142bda6e36563f5719e5b0fa37d",
		"683cafd122632142bda6e36563f5719e5b0fa37d 1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3",
		"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3 02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d",
		"a94fb112d1f2e70f55851c6c569916a9e31caee1 2716762a5213f5fe2576d2a52d1182282704004c",
	})

	expectedGraph := map[string][]string{
		"9ad62c7ec68e377b41a8b8dd846e573b76634172": {"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3", "683cafd122632142bda6e36563f5719e5b0fa37d"},
		"683cafd122632142bda6e36563f5719e5b0fa37d": {"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3"},
		"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3": {"02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d"},
		"a94fb112d1f2e70f55851c6c569916a9e31caee1": {"2716762a5213f5fe2576d2a52d1182282704004c"},
		"02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d": {},
		"2716762a5213f5fe2576d2a52d1182282704004c": {},
	}
	if diff := cmp.Diff(expectedGraph, graph.Graph()); diff != "" {
		t.Errorf("unexpected commit mapping (-want +got):\n%s", diff)
	}

	expectedOrder := []string{
		"2716762a5213f5fe2576d2a52d1182282704004c",
		"02f41985f46b400b7a673c3dfb6bab8fd1ac6a6d",
		"a94fb112d1f2e70f55851c6c569916a9e31caee1",
		"1afa9c06d8bb8b2c5746e539ed4eb80c23b21db3",
		"683cafd122632142bda6e36563f5719e5b0fa37d",
		"9ad62c7ec68e377b41a8b8dd846e573b76634172",
	}
	if diff := cmp.Diff(expectedOrder, graph.Order()); diff != "" {
		t.Errorf("unexpected commit order (-want +got):\n%s", diff)
	}
}

func TestParseRefDescriptions(t *testing.T) {
	refDescriptions, err := parseRefDescriptions([]string{
		"66a7ac584740245fc523da443a3f540a52f8af72:refs/heads/bl/symbols: :2021-01-18T16:46:51-08:00",
		"58537c06cf7ba8a562a3f5208fb7a8efbc971d0e:refs/heads/bl/symbols-2: :2021-02-24T06:21:20-08:00",
		"a40716031ae97ee7c5cdf1dec913567a4a7c50c8:refs/heads/ef/wtf: :2021-02-10T10:50:08-06:00",
		"e2e283fdaf6ea4a419cdbad142bbfd4b730080f8:refs/heads/garo/go-and-typescript-lsif-indexing: :2020-04-29T16:45:46+00:00",
		"c485d92c3d2065041bf29b3fe0b55ffac7e66b2a:refs/heads/garo/index-specific-files: :2021-03-01T13:09:42-08:00",
		"ce30aee6cc56f39d0ac6fee03c4c151c08a8cd2e:refs/heads/master:*:2021-06-16T11:51:09-07:00",
		"ec5cfc8ab33370c698273b1a097af73ea289c92b:refs/heads/nsc/bump-go-version: :2021-03-12T22:33:17+00:00",
		"22b2c4f734f62060cae69da856fe3854defdcc87:refs/heads/nsc/markupcontent: :2021-05-03T23:50:02+01:00",
		"9df3358a18792fa9dbd40d506f2e0ad23fc11ee8:refs/heads/nsc/random: :2021-02-10T16:29:06+00:00",
		"a02b85b63345a1406d7a19727f7a5472c976e053:refs/heads/sg/document-symbols: :2021-04-08T15:33:03-07:00",
		"234b0a484519129b251164ecb0674ec27d154d2f:refs/heads/symbols: :2021-01-01T22:51:55-08:00",
		"c165bfff52e9d4f87891bba497e3b70fea144d89:refs/tags/v0.10.0: :2020-08-04T08:23:30-05:00",
		"f73ee8ed601efea74f3b734eeb073307e1615606:refs/tags/v0.5.1: :2020-04-16T16:06:21-04:00",
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347:refs/tags/v0.5.2: :2020-04-16T16:20:26-04:00",
		"7886287b8758d1baf19cf7b8253856128369a2a7:refs/tags/v0.5.3: :2020-04-16T16:55:58-04:00",
		"b69f89473bbcc04dc52cafaf6baa504e34791f5a:refs/tags/v0.6.0: :2020-04-20T12:10:49-04:00",
		"172b7fcf8b8c49b37b231693433586c2bfd1619e:refs/tags/v0.7.0: :2020-04-20T12:37:36-04:00",
		"5bc35c78fb5fb388891ca944cd12d85fd6dede95:refs/tags/v0.8.0: :2020-05-05T12:53:18-05:00",
		"14faa49ef098df9488536ca3c9b26d79e6bec4d6:refs/tags/v0.9.0: :2020-07-14T14:26:40-05:00",
		"0a82af8b6914d8c81326eee5f3a7e1d1106547f1:refs/tags/v1.0.0: :2020-08-19T19:33:39-05:00",
		"262defb72b96261a7d56b000d438c5c7ec6d0f3e:refs/tags/v1.1.0: :2020-08-21T14:15:44-05:00",
		"806b96eb544e7e632a617c26402eccee6d67faed:refs/tags/v1.1.1: :2020-08-21T16:02:35-05:00",
		"5d8865d6feacb4fce3313cade2c61dc29c6271e6:refs/tags/v1.1.2: :2020-08-22T13:45:26-05:00",
		"8c45a5635cf0a4968cc8c9dac2d61c388b53251e:refs/tags/v1.1.3: :2020-08-25T10:10:46-05:00",
		"fc212da31ce157ef0795e934381509c5a50654f6:refs/tags/v1.1.4: :2020-08-26T14:02:47-05:00",
		"4fd8b2c3522df32ffc8be983d42c3a504cc75fbc:refs/tags/v1.2.0: :2020-09-07T09:52:43-05:00",
		"9741f54aa0f14be1103b00c89406393ea4d8a08a:refs/tags/v1.3.0: :2021-02-10T23:21:31+00:00",
		"b358977103d2d66e2a3fc5f8081075c2834c4936:refs/tags/v1.3.1: :2021-02-24T20:16:45+00:00",
		"2882ad236da4b649b4c1259d815bf1a378e3b92f:refs/tags/v1.4.0: :2021-05-13T10:41:02-05:00",
		"340b84452286c18000afad9b140a32212a82840a:refs/tags/v1.5.0: :2021-05-20T18:41:41-05:00",
	})
	if err != nil {
		t.Fatalf("unexpected error parsing ref descriptions: %s", err)
	}

	mustParseDate := func(s string) time.Time {
		date, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatalf("unexpected error parsing date string: %s", err)
		}

		return date
	}

	makeBranch := func(commit, name, createdDate string, isDefaultBranch bool) RefDescription {
		return RefDescription{Commit: commit, Name: name, Type: RefTypeBranch, IsDefaultBranch: isDefaultBranch, CreatedDate: mustParseDate(createdDate)}
	}

	makeTag := func(commit, name, createdDate string) RefDescription {
		return RefDescription{Commit: commit, Name: name, Type: RefTypeTag, IsDefaultBranch: false, CreatedDate: mustParseDate(createdDate)}
	}

	expectedRefDescriptions := []RefDescription{
		makeBranch("66a7ac584740245fc523da443a3f540a52f8af72", "bl/symbols", "2021-01-18T16:46:51-08:00", false),
		makeBranch("58537c06cf7ba8a562a3f5208fb7a8efbc971d0e", "bl/symbols-2", "2021-02-24T06:21:20-08:00", false),
		makeBranch("a40716031ae97ee7c5cdf1dec913567a4a7c50c8", "ef/wtf", "2021-02-10T10:50:08-06:00", false),
		makeBranch("e2e283fdaf6ea4a419cdbad142bbfd4b730080f8", "garo/go-and-typescript-lsif-indexing", "2020-04-29T16:45:46+00:00", false),
		makeBranch("c485d92c3d2065041bf29b3fe0b55ffac7e66b2a", "garo/index-specific-files", "2021-03-01T13:09:42-08:00", false),
		makeBranch("ce30aee6cc56f39d0ac6fee03c4c151c08a8cd2e", "master", "2021-06-16T11:51:09-07:00", true),
		makeBranch("ec5cfc8ab33370c698273b1a097af73ea289c92b", "nsc/bump-go-version", "2021-03-12T22:33:17+00:00", false),
		makeBranch("22b2c4f734f62060cae69da856fe3854defdcc87", "nsc/markupcontent", "2021-05-03T23:50:02+01:00", false),
		makeBranch("9df3358a18792fa9dbd40d506f2e0ad23fc11ee8", "nsc/random", "2021-02-10T16:29:06+00:00", false),
		makeBranch("a02b85b63345a1406d7a19727f7a5472c976e053", "sg/document-symbols", "2021-04-08T15:33:03-07:00", false),
		makeBranch("234b0a484519129b251164ecb0674ec27d154d2f", "symbols", "2021-01-01T22:51:55-08:00", false),
		makeTag("c165bfff52e9d4f87891bba497e3b70fea144d89", "v0.10.0", "2020-08-04T08:23:30-05:00"),
		makeTag("f73ee8ed601efea74f3b734eeb073307e1615606", "v0.5.1", "2020-04-16T16:06:21-04:00"),
		makeTag("6057f7ed8d331c82030c713b650fc8fd2c0c2347", "v0.5.2", "2020-04-16T16:20:26-04:00"),
		makeTag("7886287b8758d1baf19cf7b8253856128369a2a7", "v0.5.3", "2020-04-16T16:55:58-04:00"),
		makeTag("b69f89473bbcc04dc52cafaf6baa504e34791f5a", "v0.6.0", "2020-04-20T12:10:49-04:00"),
		makeTag("172b7fcf8b8c49b37b231693433586c2bfd1619e", "v0.7.0", "2020-04-20T12:37:36-04:00"),
		makeTag("5bc35c78fb5fb388891ca944cd12d85fd6dede95", "v0.8.0", "2020-05-05T12:53:18-05:00"),
		makeTag("14faa49ef098df9488536ca3c9b26d79e6bec4d6", "v0.9.0", "2020-07-14T14:26:40-05:00"),
		makeTag("0a82af8b6914d8c81326eee5f3a7e1d1106547f1", "v1.0.0", "2020-08-19T19:33:39-05:00"),
		makeTag("262defb72b96261a7d56b000d438c5c7ec6d0f3e", "v1.1.0", "2020-08-21T14:15:44-05:00"),
		makeTag("806b96eb544e7e632a617c26402eccee6d67faed", "v1.1.1", "2020-08-21T16:02:35-05:00"),
		makeTag("5d8865d6feacb4fce3313cade2c61dc29c6271e6", "v1.1.2", "2020-08-22T13:45:26-05:00"),
		makeTag("8c45a5635cf0a4968cc8c9dac2d61c388b53251e", "v1.1.3", "2020-08-25T10:10:46-05:00"),
		makeTag("fc212da31ce157ef0795e934381509c5a50654f6", "v1.1.4", "2020-08-26T14:02:47-05:00"),
		makeTag("4fd8b2c3522df32ffc8be983d42c3a504cc75fbc", "v1.2.0", "2020-09-07T09:52:43-05:00"),
		makeTag("9741f54aa0f14be1103b00c89406393ea4d8a08a", "v1.3.0", "2021-02-10T23:21:31+00:00"),
		makeTag("b358977103d2d66e2a3fc5f8081075c2834c4936", "v1.3.1", "2021-02-24T20:16:45+00:00"),
		makeTag("2882ad236da4b649b4c1259d815bf1a378e3b92f", "v1.4.0", "2021-05-13T10:41:02-05:00"),
		makeTag("340b84452286c18000afad9b140a32212a82840a", "v1.5.0", "2021-05-20T18:41:41-05:00"),
	}
	if diff := cmp.Diff(expectedRefDescriptions, refDescriptions); diff != "" {
		t.Errorf("unexpected ref descriptions (-want +got):\n%s", diff)
	}
}
