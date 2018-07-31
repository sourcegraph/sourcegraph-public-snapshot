CREATE INDEX idx_globaldep_package ON global_dep ((dep_data ->> 'package' COLLATE "C"));
CREATE INDEX idx_globaldep_depth ON global_dep ((dep_data ->> 'depth' COLLATE "C"));
