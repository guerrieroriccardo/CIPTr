CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Default hostname format: {short_code}{NNN} e.g. NB001, SRV001
INSERT INTO settings (key, value) VALUES
    ('hostname_prefix_source', 'short_code'),
    ('hostname_prefix_position', 'before'),
    ('hostname_num_digits', '3')
ON CONFLICT (key) DO NOTHING;
