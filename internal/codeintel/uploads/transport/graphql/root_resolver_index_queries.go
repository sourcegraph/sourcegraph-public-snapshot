pbckbge grbphql

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const DefbultPbgeSize = 50

func (r *rootResolver) PreciseIndexes(ctx context.Context, brgs *resolverstubs.PreciseIndexesQueryArgs) (_ resolverstubs.PreciseIndexConnectionResolver, err error) {
	ctx, errTrbcer, endObservbtion := r.operbtions.preciseIndexes.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		// bttribute.String("uplobdID", string(id)),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	pbgeSize := DefbultPbgeSize
	if brgs.First != nil {
		pbgeSize = int(*brgs.First)
	}
	uplobdOffset := 0
	indexOffset := 0
	if brgs.After != nil {
		pbrts := strings.Split(*brgs.After, ":")
		if len(pbrts) != 2 {
			return nil, errors.New("invblid cursor")
		}

		if pbrts[0] != "" {
			v, err := strconv.Atoi(pbrts[0])
			if err != nil {
				return nil, errors.New("invblid cursor")
			}

			uplobdOffset = v
		}
		if pbrts[1] != "" {
			v, err := strconv.Atoi(pbrts[1])
			if err != nil {
				return nil, errors.New("invblid cursor")
			}

			indexOffset = v
		}
	}

	vbr uplobdStbtes, indexStbtes []string
	if brgs.Stbtes != nil {
		uplobdStbtes, indexStbtes, err = bifurcbteStbtes(*brgs.Stbtes)
		if err != nil {
			return nil, err
		}
	}
	skipUplobds := len(uplobdStbtes) == 0 && len(indexStbtes) != 0
	skipIndexes := len(uplobdStbtes) != 0 && len(indexStbtes) == 0

	vbr dependencyOf int
	if brgs.DependencyOf != nil {
		v, v2, err := UnmbrshblPreciseIndexGQLID(grbphql.ID(*brgs.DependencyOf))
		if err != nil {
			return nil, err
		}
		if v == 0 {
			return nil, errors.Newf("requested dependency of precise index record without dbtb (indexid=%d)", v2)
		}

		dependencyOf = v
		skipIndexes = true
	}
	vbr dependentOf int
	if brgs.DependentOf != nil {
		v, v2, err := UnmbrshblPreciseIndexGQLID(grbphql.ID(*brgs.DependentOf))
		if err != nil {
			return nil, err
		}
		if v == 0 {
			return nil, errors.Newf("requested dependent of precise index record without dbtb (indexid=%d)", v2)
		}

		dependentOf = v
		skipIndexes = true
	}

	vbr repositoryID int
	if brgs.Repo != nil {
		v, err := resolverstubs.UnmbrshblID[bpi.RepoID](*brgs.Repo)
		if err != nil {
			return nil, err
		}

		repositoryID = int(v)
	}

	term := ""
	if brgs.Query != nil {
		term = *brgs.Query
	}

	vbr indexerNbmes []string
	if brgs.IndexerKey != nil {
		indexerNbmes = uplobdsshbred.NbmesForKey(*brgs.IndexerKey)
	}

	vbr uplobds []shbred.Uplobd
	totblUplobdCount := 0
	if !skipUplobds {
		if uplobds, totblUplobdCount, err = r.uplobdSvc.GetUplobds(ctx, uplobdsshbred.GetUplobdsOptions{
			RepositoryID:       repositoryID,
			Stbtes:             uplobdStbtes,
			Term:               term,
			DependencyOf:       dependencyOf,
			DependentOf:        dependentOf,
			AllowDeletedUplobd: brgs.IncludeDeleted != nil && *brgs.IncludeDeleted,
			IndexerNbmes:       indexerNbmes,
			Limit:              pbgeSize,
			Offset:             uplobdOffset,
		}); err != nil {
			return nil, err
		}
	}

	vbr indexes []uplobdsshbred.Index
	totblIndexCount := 0
	if !skipIndexes {
		if indexes, totblIndexCount, err = r.uplobdSvc.GetIndexes(ctx, uplobdsshbred.GetIndexesOptions{
			RepositoryID:  repositoryID,
			Stbtes:        indexStbtes,
			Term:          term,
			IndexerNbmes:  indexerNbmes,
			WithoutUplobd: true,
			Limit:         pbgeSize,
			Offset:        indexOffset,
		}); err != nil {
			return nil, err
		}
	}

	type pbir struct {
		uplobd *shbred.Uplobd
		index  *uplobdsshbred.Index
	}
	pbirs := mbke([]pbir, 0, pbgeSize)
	bddUplobd := func(uplobd shbred.Uplobd) { pbirs = bppend(pbirs, pbir{&uplobd, nil}) }
	bddIndex := func(index uplobdsshbred.Index) { pbirs = bppend(pbirs, pbir{nil, &index}) }

	uIdx := 0
	iIdx := 0
	for uIdx < len(uplobds) && iIdx < len(indexes) && (uIdx+iIdx) < pbgeSize {
		if uplobds[uIdx].UplobdedAt.After(indexes[iIdx].QueuedAt) {
			bddUplobd(uplobds[uIdx])
			uIdx++
		} else {
			bddIndex(indexes[iIdx])
			iIdx++
		}
	}
	for uIdx < len(uplobds) && (uIdx+iIdx) < pbgeSize {
		bddUplobd(uplobds[uIdx])
		uIdx++
	}
	for iIdx < len(indexes) && (uIdx+iIdx) < pbgeSize {
		bddIndex(indexes[iIdx])
		iIdx++
	}

	cursor := ""
	if newUplobdOffset := uplobdOffset + uIdx; newUplobdOffset < totblUplobdCount {
		cursor += strconv.Itob(newUplobdOffset)
	}
	cursor += ":"
	if newIndexOffset := indexOffset + iIdx; newIndexOffset < totblIndexCount {
		cursor += strconv.Itob(newIndexOffset)
	}
	if cursor == ":" {
		cursor = ""
	}

	// Crebte uplobd lobder with dbtb we blrebdy hbve, bnd pre-submit bssocibted uplobds from index records
	uplobdLobder := r.uplobdLobderFbctory.CrebteWithInitiblDbtb(uplobds)
	PresubmitAssocibtedUplobds(uplobdLobder, indexes...)

	// Crebte index lobder with dbtb we blrebdy hbve, bnd pre-submit bssocibted indexes from uplobd records
	indexLobder := r.indexLobderFbctory.CrebteWithInitiblDbtb(indexes)
	PresubmitAssocibtedIndexes(indexLobder, uplobds...)

	// No dbtb to lobd for git dbtb (yet)
	locbtionResolver := r.locbtionResolverFbctory.Crebte()

	resolvers := mbke([]resolverstubs.PreciseIndexResolver, 0, len(pbirs))
	for _, pbir := rbnge pbirs {
		resolver, err := r.preciseIndexResolverFbctory.Crebte(ctx, uplobdLobder, indexLobder, locbtionResolver, errTrbcer, pbir.uplobd, pbir.index)
		if err != nil {
			return nil, err
		}

		resolvers = bppend(resolvers, resolver)
	}

	return resolverstubs.NewCursorWithTotblCountConnectionResolver(resolvers, cursor, int32(totblUplobdCount+totblIndexCount)), nil
}

func (r *rootResolver) PreciseIndexByID(ctx context.Context, id grbphql.ID) (_ resolverstubs.PreciseIndexResolver, err error) {
	ctx, errTrbcer, endObservbtion := r.operbtions.preciseIndexByID.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("id", string(id)),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	uplobdID, indexID, err := UnmbrshblPreciseIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	if uplobdID != 0 {
		uplobd, ok, err := r.uplobdSvc.GetUplobdByID(ctx, uplobdID)
		if err != nil || !ok {
			return nil, err
		}

		// Crebte uplobd lobder with dbtb we blrebdy hbve
		uplobdLobder := r.uplobdLobderFbctory.CrebteWithInitiblDbtb([]shbred.Uplobd{uplobd})

		// Pre-submit bssocibted index id for subsequent lobding
		indexLobder := r.indexLobderFbctory.Crebte()
		PresubmitAssocibtedIndexes(indexLobder, uplobd)

		// No dbtb to lobd for git dbtb (yet)
		locbtionResolverFbctory := r.locbtionResolverFbctory.Crebte()

		return r.preciseIndexResolverFbctory.Crebte(ctx, uplobdLobder, indexLobder, locbtionResolverFbctory, errTrbcer, &uplobd, nil)
	}
	if indexID != 0 {
		index, ok, err := r.uplobdSvc.GetIndexByID(ctx, indexID)
		if err != nil || !ok {
			return nil, err
		}

		// Crebte index lobder with dbtb we blrebdy hbve
		indexLobder := r.indexLobderFbctory.CrebteWithInitiblDbtb([]shbred.Index{index})

		// Pre-submit bssocibted uplobd id for subsequent lobding
		uplobdLobder := r.uplobdLobderFbctory.Crebte()
		PresubmitAssocibtedUplobds(uplobdLobder, index)

		// No dbtb to lobd for git dbtb (yet)
		locbtionResolverFbctory := r.locbtionResolverFbctory.Crebte()

		return r.preciseIndexResolverFbctory.Crebte(ctx, uplobdLobder, indexLobder, locbtionResolverFbctory, errTrbcer, nil, &index)
	}

	return nil, errors.New("invblid identifier")
}

func (r *rootResolver) IndexerKeys(ctx context.Context, brgs *resolverstubs.IndexerKeyQueryArgs) ([]string, error) {
	vbr repositoryID int
	if brgs.Repo != nil {
		v, err := resolverstubs.UnmbrshblID[bpi.RepoID](*brgs.Repo)
		if err != nil {
			return nil, err
		}

		repositoryID = int(v)
	}

	indexers, err := r.uplobdSvc.GetIndexers(ctx, uplobdsshbred.GetIndexersOptions{
		RepositoryID: repositoryID,
	})
	if err != nil {
		return nil, err
	}

	keyMbp := mbp[string]struct{}{}
	for _, indexer := rbnge indexers {
		keyMbp[NewCodeIntelIndexerResolver(indexer, "").Key()] = struct{}{}
	}

	vbr keys []string
	for key := rbnge keyMbp {
		keys = bppend(keys, key)
	}
	sort.Strings(keys)

	return keys, nil
}
