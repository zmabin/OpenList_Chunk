<div align="center">
  <img src="https://raw.githubusercontent.com/OpenListTeam/Logo/main/logo.svg" width="128" height="128" alt="logo" />

  <h1>OpenList_Chunk</h1>

  <p><em>Verbeterde fork van OpenList — Omzeil CDN upload limieten met chunked upload ondersteuning</em></p>

  <a href="https://github.com/zmabin/OpenList_Chunk/actions?query=workflow%3ABuild"><img src="https://img.shields.io/github/actions/workflow/status/zmabin/OpenList_Chunk/build.yml?branch=main" alt="Build status" /></a>
  <a href="https://github.com/zmabin/OpenList_Chunk/releases"><img src="https://img.shields.io/github/v/release/zmabin/OpenList_Chunk" alt="latest version" /></a>
  <a href="https://github.com/zmabin/OpenList_Chunk/blob/main/LICENSE"><img src="https://img.shields.io/github/license/zmabin/OpenList_Chunk" alt="License" /></a>
</div>

---

[English](./README.md) | [中文](./README_cn.md) | [日本語](./README_ja.md) | **Nederlands**

---

## Overzicht

**OpenList_Chunk** is een verbeterde fork van [OpenList](https://github.com/OpenListTeam/OpenList) die de upload logica herstructureert terwijl alle originele datastructuren intact blijven.

**Kern doel: Omzeil upload limieten opgelegd door CDN reverse proxies (bijv. Cloudflare Free plan limiteert enkele verzoeken tot 100MB).**

**Direct vervanging — geen gedoe.**

---

## Kern Wijzigingen: CDN Limieten Omzeilen

Dit project implementeert twee verschillende mechanismen om CDN upload body limieten te omzeilen.

### 1. Form Chunked Upload

Een traditioneel hoog-compatibel chunking mechanisme gebaseerd op **"sessiebeheer + schijfcache + streaming merge"**.

- **Workflow**:
  1. **Init sessie**: Frontend roept `POST /api/fs/put/chunk/init` aan, backend genereert een unieke `upload_id` en maakt een sessiebestand aan.
  2. **Upload chunks**: Elke chunk wordt als `multipart/form-data` verzonden naar `PUT /api/fs/put/chunk` met `upload_id` en `index`.
  3. **CRC32 verificatie**: Server berekent CRC32 voor elke chunk en vergelijkt met de `X-Chunk-CRC32` header van de client.
  4. **Virtuele merge**: Na het uploaden van alle chunks roept de frontend `POST /api/fs/put/chunk/merge` aan. Backend gebruikt `io.MultiReader` om alle tijdelijke bestanden sequentieel te lezen zonder schijfkopie, direct streamend naar de opslag backend.
  5. **Automatische opruiming**: Tijdelijke chunk directory wordt verwijderd na merge.

- **Voordelen**: Hoge compatibiliteit, CRC32 integriteitsverificatie.
- **Beveiliging**: Elke sessie is gebonden aan de identiteit van de uploader.

### 2. Stream Chunking

Ontworpen voor maximale prestaties en minimaal resourcegebruik. Kernconcept: **"zero-copy pipe"**.

- **Workflow**:
  1. **Frontend streaming**: Frontend splitst het bestand logisch en stuurt `Raw Binary` via `PUT` met `Content-Range` headers.
  2. **io.Pipe bridge**: Bij de eerste chunk maakt de backend een zero-buffer pipe (`io.Pipe`) aan en start onmiddellijk de opslag driver upload taak die vanuit de pipe leest.
  3. **Zero-copy flow**: Volgende chunks schrijven naar dezelfde pipe. Data stroomt direct van "frontend verzoek" via "server geheugen" naar "cloud opslag".
  4. **Automatische voltooiing**: Na de laatste chunk wordt de pipe gesloten en is de upload voltooid.

- **Voordelen**:
  - **Geen schijfgebruik**: Geen tijdelijke chunks, geen schijf merge.
  - **Minimaal geheugen**: Door pipe back-pressure blijft geheugen op KB-niveau.
  - **Hoge prestaties**: Direct streaming zonder I/O bottleneck.
- **Let op**: Server fungeert als sync pipe; trage cloud snelheden zullen back-pressure uitoefenen op de client via TCP.

---

## Route Wijzigingen

| Route | Methode | Functie | Auth |
|-------|---------|---------|------|
| `/api/fs/put/chunk/init` | POST | Initialiseer chunk sessie | `FsUp` middleware |
| `/api/fs/put/chunk` | PUT | Upload een chunk | `FsUp` + rate limiet |
| `/api/fs/put/chunk/merge` | POST | Merge chunks en upload | `FsUp` + rate limiet |
| `/api/fs/put` | PUT | Stream upload (ondersteunt Content-Range) | `FsUp` + rate limiet |

---

## Implementatie Gids

### Directe Vervanging (Volledig Compatibel met OpenList Data)

1. Stop je OpenList service
2. Backup het originele `openlist` binary
3. Vervang met het gecompileerde `openlist` binary
4. Start de service

```bash
systemctl stop openlist
cp openlist /opt/openlist/openlist
chmod +x /opt/openlist/openlist
systemctl start openlist
```

### Bouw vanuit Broncode

```bash
git clone https://github.com/zmabin/OpenList_Chunk.git
cd OpenList_Chunk

# Download frontend assets
bash build.sh dev web

# Bouw (Linux)
export CGO_ENABLED=0
go build -o openlist -tags=jsoniter -ldflags="-s -w" .

# Bouw (Windows)
set CGO_ENABLED=0
go build -o openlist.exe -tags=jsoniter -ldflags="-s -w" .
```

### Docker

```bash
# ghcr.io
docker pull ghcr.io/zmabin/openlist-chunk:latest
docker run -d -p 5244:5244 -v ./data:/opt/openlist/data ghcr.io/zmabin/openlist-chunk:latest

# Docker Hub
docker pull zmabin/openlist-chunk:latest
docker run -d -p 5244:5244 -v ./data:/opt/openlist/data zmabin/openlist-chunk:latest
```

### Nginx Proxy Config

Zie `conf.d/openlist.conf` voor de volledige configuratie. Belangrijkste instellingen:

```nginx
client_max_body_size 102400m;       # 100GB max upload
proxy_request_buffering off;         # Schakel request buffering uit (vereist voor streaming)
proxy_send_timeout 86400s;           # 24-uur timeout
```

---

## Instellingen

| Sleutel | Type | Standaard | Beschrijving |
|---------|------|-----------|-------------|
| `chunked_upload_mode` | Select | `auto` | Chunk modus: `auto` / `disabled` |
| `chunked_upload_chunk_size` | Number | `95` | Chunk drempelwaarde (MB), bestanden groter dan dit worden automatisch ge-chunked |

---

## Roadmap

- [x] **Form Chunked Upload**: Sessie-gebaseerde multipart chunk + streaming merge
- [x] **Stream Chunking**: Content-Range gebaseerde zero-copy pipe chunking
- [x] **Geplande Storage Her-login**: Per-storage keep-alive via geforceerde wachtwoord herauthenticatie

---

## Met Dank aan

Dit project verwijst naar en bouwt voort op het werk van de volgende uitstekende projecten:

- Dank aan [LusiyAvA/openlist-chunk](https://github.com/LusiyAvA/openlist-chunk) voor de kernideeen en implementatie referentie voor chunked upload
- Dank aan [OpenListTeam/OpenList](https://github.com/OpenListTeam/OpenList) voor het leveren van een stabiel en betrouwbaar fundamenteel raamwerk

---

## Ondersteuning

Als dit project je helpt, overweeg dan een Star!

Een bug gevonden of een suggestie? Open gerust een [Issue](https://github.com/zmabin/OpenList_Chunk/issues).
