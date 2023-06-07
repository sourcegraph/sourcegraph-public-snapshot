INSERT INTO own_signal_configurations (name, enabled, description)
VALUES (
        'analytics',
        FALSE,
        'Indexes ownership data to present in aggregated views like Admin > Analytics > Own and Repo > Ownership'
    ) ON CONFLICT DO NOTHING;