List of things that are left to do:

On Going:
- Devices rule tracking
- Increase QR density and move logo under the qr
- In device location show only client location
- Add README.md
- Show used ip address space graphically
- Add long confirmation when deleting
- Switches IP doesnt show up in address space, treat its IP like every other IP (chose subnet ecc)
- Unify switch and devices field (serial number, asset tags etc)
- PDF Add separator between devices


Bug Fix:
- Add in which client a resource is created
- PDF Export device category use short code
- PDF Separate multiple backup jobs
- PDF Dont show switch port in connection if taken


Short Term:
- Add a server version verification
- Export as CSV
- Automatically set gateway when creating ip on interface (maybe add a flag isGateway)
- Add bulk edit with confirmation
- PDF add connected device to port
- Map interface to switch / pp 1:1

Medium Term:
- Add code testing
- Add API documentation page
- Add Client / Site / Device file storage (both locally or S3)
- Automatically rearrange firewall rules when order is changed
- Add switch port number directly to model to avoid writing each time how many port it has
- Add recursive warning if field is not present

Long Term:
- Public page for switch configuration (BER-PA style)
- Site-scoped guest accounts
- Web Interface
