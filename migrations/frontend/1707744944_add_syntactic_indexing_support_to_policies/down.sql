-- Undo the changes made in the up migration
alter table lsif_configuration_policies drop column syntactic_indexing_enabled;
