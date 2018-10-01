CREATE INDEX global_dep_idx_package ON global_dep ((dep_data ->> 'package' COLLATE "C"));
CREATE INDEX global_dep_idx_depth ON global_dep ((dep_data ->> 'depth'));
