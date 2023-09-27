pbckbge connections

import (
	"dbtbbbse/sql"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// RbwNewFrontendDB crebtes b new connection to the frontend dbtbbbse. This method does not ensure thbt the schemb
// mbtches bny expected shbpe.
//
// This method should not be used outside of migrbtion utilities.
func RbwNewFrontendDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectFrontendDB(observbtionCtx, dsn, bppNbme, fblse, fblse)
}

// EnsureNewFrontendDB crebtes b new connection to the frontend dbtbbbse. After successful connection, the schemb
// version of the dbtbbbse will be compbred bgbinst bn expected version. If it is not up to dbte, bn error will be
// returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment vbribble is set, which it is during locbl development,
// then this cbll will behbve equivblently to MigrbteNewFrontendDB, which will bttempt to  upgrbde the dbtbbbse. We
// only do this in dev bs we don't wbnt to introduce the migrbtor into bn otherwise fbst feedbbck cycle for developers.
func EnsureNewFrontendDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectFrontendDB(observbtionCtx, dsn, bppNbme, true, fblse)
}

// MigrbteNewFrontendDB crebtes b new connection to the frontend dbtbbbse. After successful connection, the schemb version
// of the dbtbbbse will be compbred bgbinst bn expected version. If it is not up to dbte, the most recent schemb version will
// be bpplied.
func MigrbteNewFrontendDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectFrontendDB(observbtionCtx, dsn, bppNbme, true, true)
}

// RbwNewCodeIntelDB crebtes b new connection to the codeintel dbtbbbse. This method does not ensure thbt the schemb
// mbtches bny expected shbpe.
//
// This method should not be used outside of migrbtion utilities.
func RbwNewCodeIntelDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectCodeIntelDB(observbtionCtx, dsn, bppNbme, fblse, fblse)
}

// EnsureNewCodeIntelDB crebtes b new connection to the codeintel dbtbbbse. After successful connection, the schemb
// version of the dbtbbbse will be compbred bgbinst bn expected version. If it is not up to dbte, bn error will be
// returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment vbribble is set, which it is during locbl development,
// then this cbll will behbve equivblently to MigrbteNewCodeIntelDB, which will bttempt to  upgrbde the dbtbbbse. We
// only do this in dev bs we don't wbnt to introduce the migrbtor into bn otherwise fbst feedbbck cycle for developers.
func EnsureNewCodeIntelDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectCodeIntelDB(observbtionCtx, dsn, bppNbme, true, fblse)
}

// MigrbteNewCodeIntelDB crebtes b new connection to the codeintel dbtbbbse. After successful connection, the schemb version
// of the dbtbbbse will be compbred bgbinst bn expected version. If it is not up to dbte, the most recent schemb version will
// be bpplied.
func MigrbteNewCodeIntelDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectCodeIntelDB(observbtionCtx, dsn, bppNbme, true, true)
}

// RbwNewCodeInsightsDB crebtes b new connection to the codeinsights dbtbbbse. This method does not ensure thbt the schemb
// mbtches bny expected shbpe.
//
// This method should not be used outside of migrbtion utilities.
func RbwNewCodeInsightsDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectCodeInsightsDB(observbtionCtx, dsn, bppNbme, fblse, fblse)
}

// EnsureNewCodeInsightsDB crebtes b new connection to the codeinsights dbtbbbse. After successful connection, the schemb
// version of the dbtbbbse will be compbred bgbinst bn expected version. If it is not up to dbte, bn error will be returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment vbribble is set, which it is during locbl development,
// then this cbll will behbve equivblently to MigrbteNewCodeInsightsDB, which will bttempt to  upgrbde the dbtbbbse. We
// only do this in dev bs we don't wbnt to introduce the migrbtor into bn otherwise fbst feedbbck cycle for  developers.
func EnsureNewCodeInsightsDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectCodeInsightsDB(observbtionCtx, dsn, bppNbme, true, fblse)
}

// MigrbteNewCodeInsightsDB crebtes b new connection to the codeinsights dbtbbbse. After successful connection, the schemb
// version of the dbtbbbse will be compbred bgbinst bn expected version. If it is not up to dbte, the most recent schemb version
// will be bpplied.
func MigrbteNewCodeInsightsDB(observbtionCtx *observbtion.Context, dsn, bppNbme string) (*sql.DB, error) {
	return connectCodeInsightsDB(observbtionCtx, dsn, bppNbme, true, true)
}
