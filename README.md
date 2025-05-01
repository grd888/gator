# Gator - RSS Feed Aggregator CLI

Gator is a command-line RSS feed aggregator that allows users to register, follow RSS feeds, and browse posts from those feeds.

## Prerequisites

Before you can use Gator, you need to have the following installed:

- **Go** (version 1.16 or later) - [Install Go](https://golang.org/doc/install)
- **PostgreSQL** - [Install PostgreSQL](https://www.postgresql.org/download/)

## Installation

To install the Gator CLI, run:

```bash
go install github.com/grd888/gator@latest
```

This will compile and install the `gator` binary to your `$GOPATH/bin` directory. Make sure this directory is in your system's `PATH` to run the command from anywhere.

## Configuration

Gator requires a configuration file to connect to your PostgreSQL database. The configuration file is stored at `~/.gatorconfig.json`.

Create this file with the following content:

```json
{
  "db_url": "postgresql://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace `username`, `password`, and `gator` with your PostgreSQL credentials and database name.

## Usage

Gator provides several commands to interact with RSS feeds:

### User Management

- **Register a new user**:
  ```
  gator register <username>
  ```

- **Login as an existing user**:
  ```
  gator login <username>
  ```

- **List all users**:
  ```
  gator users
  ```

- **Reset (delete all users)**:
  ```
  gator reset
  ```

### Feed Management

- **Add a new RSS feed**:
  ```
  gator addfeed <url>
  ```

- **List all available feeds**:
  ```
  gator feeds
  ```

- **Follow a feed**:
  ```
  gator follow <url>
  ```

- **List feeds you're following**:
  ```
  gator following
  ```

- **Unfollow a feed**:
  ```
  gator unfollow <url>
  ```

### Content Browsing

- **Browse posts from feeds you follow**:
  ```
  gator browse [limit]
  ```
  The optional `limit` parameter specifies how many posts to show (default: 2).

- **Aggregate and update feeds**:
  ```
  gator agg
  ```
  This command fetches and updates the content from RSS feeds.

## Example Workflow

1. Register a new user:
   ```
   gator register johndoe
   ```

2. Add an RSS feed:
   ```
   gator addfeed https://example.com/rss
   ```

3. Follow the feed:
   ```
   gator follow https://example.com/rss
   ```

4. Aggregate feed content:
   ```
   gator agg
   ```

5. Browse posts from followed feeds:
   ```
   gator browse 5
   ```

## Development

To contribute to Gator, clone the repository and install dependencies:

```bash
git clone https://github.com/grd888/gator.git
cd gator
go mod download
```

## License

This project is open source and available under the [MIT License](LICENSE).
