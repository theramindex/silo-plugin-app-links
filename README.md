# Silo App Links

Silo plugin that adds an Apps launcher for external services. Admins manage links in a custom admin panel. Users open links in a Silo-hosted fullscreen iframe shell or in a new tab.

## Storage

Links are stored in JSON. By default the plugin writes:

`/var/lib/continuum/plugins/silo.ramindex.app-links/app-links.json`

Override with:

`APP_LINKS_DATA_FILE=/path/to/app-links.json`

## Routes

- User app: `/app-links`
- Fullscreen iframe shell: `/app-links/open?id=<link-id>`
- Admin manager: `/app-links/admin`
