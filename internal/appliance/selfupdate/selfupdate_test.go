package selfupdate

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/schema"
	"github.com/stretchr/testify/assert"
)

type mockUpdater struct {
	calls []*schema.ComponentUpdateInformation
}

func (m *mockUpdater) Update(comp *schema.ComponentUpdateInformation) (*semver.Version, error) {
	m.calls = append(m.calls, comp)
	newVer, _ := semver.NewVersion(comp.Version)
	return newVer, nil
}

var ConfigWithoutComponents = schema.SelfUpdateDefinition{
	SelfUpdate: schema.ComponentUpdateInformation{
		Name:        "self-update",
		Version:     "1.2.3",
		DisplayName: "Self Updater",
		UpdateUrl:   "http://nowhere/blah/blah",
	},
	Components: []schema.ComponentUpdateInformation{},
}

var ConfigWithComponents = schema.SelfUpdateDefinition{
	SelfUpdate: schema.ComponentUpdateInformation{
		Name:        "self-update",
		Version:     "1.2.3",
		DisplayName: "Self Updater",
		UpdateUrl:   "http://nowhere/blah/blah",
	},
	Components: []schema.ComponentUpdateInformation{
		{
			Name:        "component-one",
			DisplayName: "Component O-N-E",
			UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
		},
		{
			Name:        "component-two",
			DisplayName: "Component T-W-O",
			UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
		},
		{
			Name:        "component-three",
			DisplayName: "Component III",
			UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
		},
	},
}

func TestSameVersion(t *testing.T) {
	mock := &mockUpdater{}
	config := ConfigWithoutComponents
	sut := selfupdater{
		currentVersion: "1.2.3",
		updater:        mock,
	}
	assert.NoError(t, sut.Start(&config))
	assert.Equal(t, 0, len(mock.calls))
}

func TestSelfSelfUpdateVersion(t *testing.T) {
	mock := &mockUpdater{}
	triedToExit := false
	config := ConfigWithoutComponents
	config.SelfUpdate.Version = "2.3.4"
	sut := selfupdater{
		currentVersion: "1.2.3",
		updater:        mock,
		exitHandler: func() {
			triedToExit = true
		},
	}
	assert.NoError(t, sut.Start(&config))
	assert.Equal(t, 1, len(mock.calls))
	assert.Equal(t, &config.SelfUpdate, mock.calls[0])
	assert.True(t, triedToExit)
}

func TestSelfSelfUpdateDoesNotUpdateComponents(t *testing.T) {
	mock := &mockUpdater{}
	triedToExit := false
	config := ConfigWithComponents
	config.SelfUpdate.Version = "2.3.4"
	sut := selfupdater{
		currentVersion: "1.2.3",
		updater:        mock,
		exitHandler: func() {
			triedToExit = true
		},
	}
	assert.NoError(t, sut.Start(&config))
	assert.Equal(t, 1, len(mock.calls)) // we will not update others
	assert.Equal(t, &config.SelfUpdate, mock.calls[0])
	assert.True(t, triedToExit)
}

func TestUpdateComponents(t *testing.T) {
	mock := &mockUpdater{}
	triedToExit := false
	config := ConfigWithComponents
	sut := selfupdater{
		currentVersion: "1.2.3",
		updater:        mock,
		exitHandler: func() {
			triedToExit = true
		},
	}
	assert.NoError(t, sut.Start(&config))
	assert.Equal(t, 3, len(mock.calls))
	assert.Equal(t, &config.Components[0], mock.calls[0])
	assert.Equal(t, &config.Components[1], mock.calls[1])
	assert.Equal(t, &config.Components[2], mock.calls[2])
	assert.False(t, triedToExit)
}
