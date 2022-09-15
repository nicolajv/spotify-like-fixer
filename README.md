# Spotify like fixer

Replaces liked songs that were reuploaded, causing the like-duplication bug. Simply un- and relikes any song with exactly the same name and album, but different ids.

Requires setting up a Spotify application, putting its client id and secret in .env and setting its redirect URI to `http://localhost:8080/callback`
