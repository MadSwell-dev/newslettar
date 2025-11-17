# Contributing to Newslettar

First off, thank you for considering contributing to Newslettar! It's people like you that make this project great.

## Code of Conduct

This project and everyone participating in it is governed by basic principles of respect and professionalism. Please be kind and courteous.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** - Include configuration snippets, logs, or screenshots
- **Describe the behavior you observed** and what you expected to see
- **Include environment details:**
  - OS and version
  - Docker version (if using Docker)
  - Go version (if building from source)
  - Newslettar version (check `version.json` or `newslettar-ctl version`)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful** to most users
- **List any similar features** in other applications if applicable

### Pull Requests

1. **Fork the repository** and create your branch from `main`:
   ```bash
   git checkout -b feature/my-amazing-feature
   ```

2. **Make your changes:**
   - Follow the existing code style
   - Add comments for complex logic
   - Update documentation if needed

3. **Test your changes:**
   ```bash
   make test
   make build
   make run
   ```

4. **Ensure code is formatted:**
   ```bash
   make fmt
   ```

5. **Commit your changes:**
   ```bash
   git commit -m "feat: add amazing feature"
   ```

   Use conventional commit messages:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `style:` - Code style changes (formatting, etc.)
   - `refactor:` - Code refactoring
   - `test:` - Adding or updating tests
   - `chore:` - Maintenance tasks

6. **Push to your fork:**
   ```bash
   git push origin feature/my-amazing-feature
   ```

7. **Open a Pull Request** with a clear title and description

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Git
- Make (optional but recommended)
- Docker (for testing Docker builds)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/newslettar.git
cd newslettar

# Build
make build

# Run locally
make run

# Run tests
make test

# Format code
make fmt
```

### Project Structure

```
newslettar/
├── main.go           # Application entry point
├── types.go          # Type definitions
├── config.go         # Configuration handling
├── api.go            # Sonarr/Radarr API client
├── trakt.go          # Trakt.tv API client
├── newsletter.go     # Newsletter generation
├── handlers.go       # HTTP handlers
├── server.go         # Server and scheduler
├── ui.go             # Web UI templates
├── utils.go          # Utility functions
├── templates/        # Email templates
└── assets/           # Static assets (logos)
```

### Coding Standards

- **Go Style:** Follow [Effective Go](https://golang.org/doc/effective_go.html)
- **Formatting:** Use `gofmt` (or `make fmt`)
- **Comments:** Add comments for exported functions and complex logic
- **Error Handling:** Always handle errors explicitly
- **Naming:** Use descriptive names, avoid abbreviations
- **Testing:** Write tests for new functionality

### Testing

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Test specific package
go test -v ./...
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build Docker image
make docker-build

# Build Debian package
make deb
```

## What to Contribute

### Good First Issues

Look for issues labeled `good first issue` - these are specifically curated for new contributors.

### Priority Areas

- **Documentation improvements** - Always welcome!
- **Bug fixes** - Help make Newslettar more stable
- **Tests** - Improve code coverage
- **Performance optimizations** - Make it faster
- **New integrations** - Plex, Jellyfin, etc.
- **UI improvements** - Better web interface
- **Email template designs** - More themes

### Features We're Looking For

- Multi-language support
- Custom email templates
- Additional media server integrations (Plex, Jellyfin)
- Discord/Slack notifications
- Web-based template editor
- Statistics and analytics dashboard
- Multiple newsletter configurations

## Documentation

When adding features or making changes:

1. **Update README.md** if user-facing behavior changes
2. **Update code comments** for complex logic
3. **Add inline documentation** for new functions
4. **Update configuration examples** if adding new settings

## Questions?

Don't hesitate to ask questions by:
- Opening a GitHub issue
- Starting a discussion in GitHub Discussions
- Emailing hello@agencefanfare.com

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Newslettar! 🎉
