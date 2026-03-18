CREATE TABLE used_vm_ids (
    id        BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    vm_id     INT NOT NULL,
    UNIQUE(client_id, vm_id)
);

-- Seed from any vm_ids already assigned to existing devices.
INSERT INTO used_vm_ids (client_id, vm_id)
SELECT s.client_id, d.vm_id
FROM devices d
JOIN sites s ON s.id = d.site_id
WHERE d.vm_id IS NOT NULL
ON CONFLICT DO NOTHING;
