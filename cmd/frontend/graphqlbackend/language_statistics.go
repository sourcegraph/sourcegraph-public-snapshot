package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"

type languageStatisticsResolver struct {
	l inventory.Lang
}

func (l *languageStatisticsResolver) Name() string {
	return l.l.Name
}

func (l *languageStatisticsResolver) TotalBytes() int32 {
	return int32(l.l.TotalBytes)
}

func (l *languageStatisticsResolver) TotalLines() int32 {
	return int32(l.l.TotalLines)
}
