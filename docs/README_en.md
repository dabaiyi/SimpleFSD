# SimpleFSD

[中文](../README.md)

An FSD (Flight Simulator Daemon) for simulated flight multiplayer, written in Go.

Supports flight plan synchronization, flight plan locking, and web-based flight plan submission.

---
[![ReleaseCard]][Release]![ReleaseDataCard]
![LastCommitCard]![BuildStateCard]
![ProjectLanguageCard]![ProjectLicense]
---

## Project Introduction

This FSD server supports out-of-the-box use. You can quickly set up an FSD server by running the server executable twice
consecutively.

The first run will error and exit automatically. This is normal behavior. The first run generates a configuration file
template in the same directory.

At this point, you can:

1. Follow the instructions in the [Configuration File Introduction](#configuration-file-introduction) to complete the
   server configuration before starting.
2. Run the server executable directly. The server will run with default configuration.

Note: The default configuration uses an SQLite database for data storage. SQLite has significant performance bottlenecks
with multi-threaded writes and stores data in a single file on disk, making it highly dependent on hard disk
performance.

> Thank you for the simple test result of 3370
> When deployed locally without considering bandwidth, sqlite can support a maximum of approximately 200 to 300 clients and there is a probability of disconnection
> When using the mysql database, it can easily run to over 400 clients and still has spare capacity ~~(because the test program is written in python, the test program has encountered a bottleneck)~~

Therefore, we recommend:

1. Do not use SQLite as a long-term database solution.
2. Do not perform high-traffic or high-pressure tests when using SQLite as the database.

If you truly lack the means to deploy a large relational database (like MySQL),

Then we recommend turning the `simulator_server` switch on, i.e., using this server as a simulator server.

Because in simulator server mode, flight crew flight plans are not written to the database; a large portion is cached in
memory, which can somewhat alleviate SQLite's poor write performance.

This project is not a pure FSD project; it also integrates an HTTP API server and a gRPC server.

Check [here](#links) for more detailed descriptions and API documentation.

## How to Use

### Usage

1. Download the corresponding build from the [Releases][Release] page.
2. Clone the entire repository and build it yourself.
3. You can also go to the [Actions][Action] page to get the latest development version (development versions may be
   unstable and contain bugs, use with caution).

### Building

1. Clone this repository.
2. Ensure you have the Go compiler installed, version >= 1.23.4.
3. Run the following command in the project root directory: `go build -x .\cmd\fsd-server\` (Windows) or
   `go build -x ./cmd/fsd-server/` (Linux).
4. \[Optional\] Compress the executable using UPX (Windows): `upx.exe -9 .\fsd-server.exe` or (Linux):
   `upx -9 ./fsd-server`.
5. After compilation completes, for Windows users, run the generated `fsd-server.exe`; for Linux users, run the
   generated `fsd-server` file.
6. The first run will create a configuration file `config.json` in the same directory as the executable. Please edit the
   configuration file and start the server again.
7. Enjoy.

### Configuration File Introduction

```json5
{
  // Debug mode, outputs extensive logs. Do not enable in production.
  "debug_mode": false,
  // Configuration file version, usually matches the software version.
  "config_version": "0.5.0",
  // Service configuration
  "server": {
    // General configuration items
    "general": {
      // Whether it is a simulator server
      // Because checking if a web-submitted plan matches the actual connection plan is needed,
      // flight plans are stored identified by user CID.
      // But simulators all use the same user CID, which causes problems:
      // incorrect simulator plans or inability to retrieve plans.
      // Set this variable to true in this case.
      // The server will then use the callsign as the identifier.
      // However, this loses the callsign matching check functionality.
      // Web plan submission is still available, just without the check.
      // Hence naming this switch the Simulator Server switch.
      "simulator_server": false,
      // Bcrypt encryption cost (number of rounds)
      "bcrypt_cost": 12
    },
    // FSD server configuration
    "fsd_server": {
      // FSD name, sent to clients connecting to the server as a MOTD message.
      "fsd_name": "Simple-Fsd",
      // FSD server listen address.
      "host": "0.0.0.0",
      // FSD server listen port.
      "port": 6809,
      // Airport data file path. If it doesn't exist, it will be downloaded from GitHub automatically.
      "airport_data_file": "data/airport.json",
      // Server flight path recording interval.
      // Meaning: Record position every N packets received from the client.
      "pos_update_points": 1,
      // FSD server heartbeat interval.
      "heartbeat_interval": "60s",
      // FSD server session expiration time.
      // Reconnecting within this time will automatically match the disconnected session.
      // Otherwise, a new session will be created.
      "session_clean_time": "40s",
      // Maximum number of worker threads, effectively the maximum number of simultaneous socket connections.
      "max_workers": 128,
      // Maximum broadcast threads, maximum threads for broadcasting messages.
      "max_broadcast_workers": 128,
      // MOTD message to send to clients.
      "motd": [
        "This is my test fsd server"
      ]
    },
    // HTTP server configuration
    "http_server": {
      // Whether to enable the HTTP server.
      "enabled": false,
      // HTTP server listen address.
      "host": "0.0.0.0",
      // HTTP server listen port.
      "port": 6810,
      // HTTP server maximum worker threads.
      "max_workers": 128,
      // Whazzup update cache time.
      "whazzup_cache_time": "15s",
      // Whazzup access URL.
      // Needs to be the external access address for Little Navmap's online flight display.
      // If you don't need Little Navmap's online flight display, you can ignore this.
      "whazzup_url_header": "http://127.0.0.1:6810",
      // Proxy type.
      // 0 Direct connection, no proxy server.
      // 1 Proxy server uses X-Forwarded-For HTTP header.
      // 2 Proxy server uses X-Real-Ip HTTP header.
      "proxy_type": 0,
      // POST request body size limit.
      // Set to an empty string to disable the limit.
      "body_limit": "10MB",
      // Server storage configuration.
      "store": {
        // Storage type. Available options:
        // 0 Local storage
        // 1 Alibaba Cloud OSS
        // 2 Tencent Cloud COS
        "store_type": 0,
        // Storage bucket region. Invalid for local storage.
        "region": "",
        // Storage bucket name. Invalid for local storage.
        "bucket": "",
        // Access ID. Invalid for local storage.
        "access_id": "",
        // Access Key. Invalid for local storage.
        "access_key": "",
        // CDN acceleration domain. Invalid for local storage.
        "cdn_domain": "",
        // Use internal network URL for uploads. Invalid for Alibaba Cloud OSS.
        "use_internal_url": false,
        // Local file save path.
        "local_store_path": "uploads",
        // Remote file save path. Invalid for local storage.
        "remote_store_path": "fsd",
        // File limits.
        "file_limit": {
          // Image file limits.
          "image_limit": {
            // Maximum allowed file size in Byte.
            "max_file_size": 5242880,
            // Allowed file extensions.
            "allowed_file_ext": [
              ".jpg",
              ".png",
              ".bmp",
              ".jpeg"
            ],
            // Storage path prefix.
            "store_prefix": "images",
            // Also keep a copy locally on the server.
            "store_in_server": false
          }
        }
      },
      "limits": {
        // API access rate limit.
        // Calculated separately per IP per endpoint.
        "rate_limit": 60,
        // API access rate limit window.
        // The rate_limit applies per rate_limit_window sliding window.
        "rate_limit_window": "1m",
        // Minimum username length.
        "username_length_min": 4,
        // Maximum username length (system max is 64).
        "username_length_max": 16,
        // Minimum email length.
        "email_length_min": 4,
        // Maximum email length (system max is 128).
        "email_length_max": 64,
        // Minimum password length.
        "password_length_min": 6,
        // Maximum password length (system max is 128).
        "password_length_max": 64,
        // Minimum CID.
        "cid_min": 1,
        // Maximum CID (system max is 2147483647).
        "cid_max": 9999,
      },
      // Email configuration.
      "email": {
        // SMTP server address.
        "host": "smtp.example.com",
        // SMTP server port.
        "port": 465,
        // Sender account.
        "username": "noreply@example.cn",
        // Sender account password or access token.
        "password": "123456",
        // Email verification code expiration time.
        "verify_expired_time": "5m",
        // Verification code resend interval.
        "send_interval": "1m",
        // Email template definitions.
        "template": {
          // Verification code template file path. Downloaded from GitHub if missing.
          "email_verify_template_file": "template/email_verify.template",
          // ATC rating change notification template file path. Downloaded from GitHub if missing.
          "atc_rating_change_template_file": "template/atc_rating_change.template",
          // Enable ATC rating change notifications.
          "enable_rating_change_email": true,
          // Flight control permission change notification template file path. Downloaded from GitHub if missing.
          "permission_change_template_file": "template/permission_change.template",
          // Enable flight control permission change notifications.
          "enable_permission_change_email": true,
          // Kicked from server notification template file path. Downloaded from GitHub if missing.
          "kicked_from_server_template_file": "template/kicked_from_server.template",
          // Enable kicked from server notifications.
          "enable_kicked_from_server_email": true
        }
      },
      // JWT configuration.
      "jwt": {
        // JWT symmetric encryption secret key.
        // PLEASE protect this key.
        // Ensure it is not known by anyone untrusted.
        // If leaked, anyone can forge administrator users.
        // A safer practice is to leave this field empty, which invalidates all previous keys on server restart.
        "secret": "123456",
        // JWT master token expiration time.
        // Recommended not more than 1 hour, as JWT tokens are stateless.
        // Long expiration times may cause security issues.
        "expires_time": "15m",
        // JWT refresh token expiration time.
        // This time is added *after* the master token expires.
        // E.g., if both are 1h, refresh token expires 2h after issuance.
        "refresh_time": "24h"
      },
      // SSL configuration.
      "ssl": {
        // Enable SSL.
        "enable": false,
        // Enable HSTS.
        "enable_hsts": false,
        // HSTS expiration time (seconds).
        "hsts_expired_time": 5184000,
        // 60 days
        // HSTS include subdomains.
        // WARNING: If not all subdomains have SSL certificates,
        // enabling this may make non-SSL domains inaccessible.
        // Do not enable if unsure.
        "include_domain": false,
        // SSL certificate file path.
        "cert_file": "",
        // SSL private key file path.
        "key_file": ""
      }
    },
    // gRPC server configuration.
    "grpc_server": {
      // Enable gRPC server.
      "enabled": false,
      // gRPC server listen address.
      "host": "0.0.0.0",
      // gRPC server listen port.
      "port": 6811,
      // gRPC server API cache time.
      "whazzup_cache_time": "15s"
    }
  },
  // Database configuration.
  "database": {
    // Database type. Supported: mysql, postgres, sqlite3.
    "type": "mysql",
    // For sqlite3: database file path and name.
    // For others: database name to use.
    "database": "go-fsd",
    // Database host address.
    "host": "localhost",
    // Database port.
    "port": 3306,
    // Database username.
    "username": "root",
    // Database password.
    "password": "123456",
    // Enable SSL for database connection.
    "enable_ssl": false,
    // Database connection pool idle timeout.
    "connect_idle_timeout": "1h",
    // Connection timeout.
    "connect_timeout": "5s",
    // Database maximum connections.
    "server_max_connections": 32
  },
  // Special permissions configuration. See `Special Permissions Configuration` chapter.
  "rating": {}
}
```

### Permissions Definition Table

#### FSD ATC Ratings Overview

| Rating Identifier | Value | Name               | Notes                                        |
|:------------------|:-----:|:-------------------|:---------------------------------------------|
| Ban               |  -1   | Banned             |                                              |
| Normal            |   0   | Normal User        | Default rating                               |
| Observer          |   1   | Observer           |                                              |
| STU1              |   2   | Delivery/Ground    |                                              |
| STU2              |   3   | Tower              |                                              |
| STU3              |   4   | Approach/Departure |                                              |
| CTR1              |   5   | Center             |                                              |
| CTR2              |   6   | Center             | Deprecated, listed for compatibility with ES |
| CTR3              |   7   | Center             |                                              |
| Instructor1       |   8   | Instructor         |                                              |
| Instructor2       |   9   | Instructor         |                                              |
| Instructor3       |  10   | Instructor         |                                              |
| Supervisor        |  11   | Supervisor         |                                              |
| Administrator     |  12   | Administrator      |                                              |

#### ATC Positions Overview

| Position Identifier | Position Code | Name     | Notes                      |
|:--------------------|:-------------:|:---------|:---------------------------|
| Pilot               |       1       | Pilot    | Connected pilots use this. |
| OBS                 |       2       | Observer |                            |
| DEL                 |       4       | Delivery |                            |
| GND                 |       8       | Ground   |                            |
| TWR                 |      16       | Tower    |                            |
| APP                 |      32       | Approach |                            |
| CTR                 |      64       | Center   |                            |
| FSS                 |      128      | FSS      |                            |

#### ATC Rating to Position Mapping

| Rating Identifier | Allowed Positions                        | Notes       |
|:------------------|:-----------------------------------------|:------------|
| Ban               | No positions allowed                     | Banned user |
| Normal            | Pilot                                    |             |
| Observer          | Pilot, OBS                               |             |
| STU1              | Pilot, OBS, DEL, GND                     |             |
| STU2              | Pilot, OBS, DEL, GND, TWR                |             |
| STU3              | Pilot, OBS, DEL, GND, TWR, APP           |             |
| CTR1              | Pilot, OBS, DEL, GND, TWR, APP, CTR      |             |
| CTR2              | Pilot, OBS, DEL, GND, TWR, APP, CTR      |             |
| CTR3              | Pilot, OBS, DEL, GND, TWR, APP, CTR, FSS |             |
| Instructor1       | Pilot, OBS, DEL, GND, TWR, APP, CTR, FSS |             |
| Instructor2       | Pilot, OBS, DEL, GND, TWR, APP, CTR, FSS |             |
| Instructor3       | Pilot, OBS, DEL, GND, TWR, APP, CTR, FSS |             |
| Supervisor        | Pilot, OBS, DEL, GND, TWR, APP, CTR, FSS |             |
| Administrator     | Pilot, OBS, DEL, GND, TWR, APP, CTR, FSS |             |

### Special Permissions Configuration

You can override the default ATC Rating to Position mapping table via the configuration file.
**WARNING!!!** This field *overrides* the default mapping table.
Do not modify this configuration unless you explicitly know what you are doing.
The configuration file field is `rating`.

```json5
{
  // Special permissions configuration
  "rating": {
    // The key is the *value* of the rating identifier you want to modify.
    // Example: Allow Normal rating to also use OBS position (Pilots connect as Observer).
    // Normal rating value is 0, so the key is "0".
    // The value is the sum of the position codes you want to permit.
    // Example: Allow pilot to connect normally (Pilot=1) and as Observer (OBS=2).
    // Value is 1 + 2 = 3.
    // If you want them to also use FSS (Please don't actually do this for Normal users!)
    // Value is 1 + 2 + 128 = 131.
    // Mappings for other ratings remain default.
    // You can also set a rating's value to 0 to prevent FSD login for that rating.
    "0": 3
    // Allow Normal users Pilot (1) and Observer (2) positions (1+2=3)
  }
}
```

## Feedback Method

If you encounter any bugs or suspected bugs while using FSD, please submit
an [Issue].
When submitting, please follow these steps:

1. For reproducible bugs:
    1. Enable `"debug_mode": true` in the [configuration file](#configuration-file-introduction) to enable log file
       output.
    2. Restart FSD and reproduce the bug.
    3. Upload the log file to the [Issue].
2. For non-reproducible bugs:
    1. Describe the bug in text as accurately as possible and submit it to the [Issue].

## Links

[Http Api Documentation][HttpApiDocs]

## Open Source License

MIT License

Copyright (c) 2025 Half_nothing

No additional terms.

## Code of Conduct

See [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md)

[ReleaseCard]: https://img.shields.io/github/v/release/Flyleague-Collection/fsd-server?style=for-the-badge&logo=github

[ReleaseDataCard]: https://img.shields.io/github/release-date/Flyleague-Collection/fsd-server?display_date=published_at&style=for-the-badge&logo=github

[LastCommitCard]: https://img.shields.io/github/last-commit/Flyleague-Collection/fsd-server?display_timestamp=committer&style=for-the-badge&logo=github

[BuildStateCard]: https://img.shields.io/github/actions/workflow/status/Flyleague-Collection/fsd-server/go-build.yml?style=for-the-badge&logo=github

[ProjectLanguageCard]: https://img.shields.io/github/languages/top/Flyleague-Collection/fsd-server?style=for-the-badge&logo=github

[ProjectLicense]: https://img.shields.io/badge/License-MIT-blue?style=for-the-badge&logo=github

[Release]: https://www.github.com/Flyleague-Collection/fsd-server/releases/latest

[Action]: https://github.com/Flyleague-Collection/fsd-server/actions/workflows/go-build.yml

[Issue]: https://github.com/Flyleague-Collection/fsd-server/issues/new

[HttpApiDocs]: https://fsd.docs.half-nothing.cn/