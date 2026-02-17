# Repository Guidelines

## Project Structure & Module Organization

- `cmd/`: entrypoints and utilities (`cmd/main`, `cmd/lambda`, import/seed/build tools).
- `internal/`: core app code (`handlers`, `services`, `db`, `models`, `views`, `auth`).
- `internal/views/*.templ`: server-rendered templates; generated `*_templ.go` files are ignored.
- `static/`: source JS, images, fonts, and CSS assets.
- `src/input.css` + `tailwind.config.js`: Tailwind source/config.
- `templates/`: CSV templates (guest imports).
- `terraform/`: infrastructure (CloudFront, DynamoDB, WAF, ACM).
- `docs/`: implementation notes and design docs.

## Build, Test, and Development Commands

- `make build`: compile local server binary (`bin/main`).
- `go test ./...`: run all Go package checks.
- `make tpl`: regenerate templ code from `.templ` files.
- `npm run styles` / `npm run styles:watch`: build/watch Tailwind CSS.
- `make scripts-dev`: copy JS to `dist/js` and minify/fingerprint scripts.
- `make dev-build`: full local asset prep (templ, metadata, image optimization, scripts).
- `make server-local`: run app with local DynamoDB endpoint.
- Local DB helpers: `make db-start`, `make db-setup`, `make db-seed`, `make db-stop`.

## Coding Style & Naming Conventions

- Go: always run `gofmt` on changed `.go` files; use idiomatic Go naming.
- Keep handlers thin; business logic in `internal/services` when reusable.
- Template helpers should stay near related `.templ` files and remain small.
- JS in `static/js` should use clear function names and avoid framework-specific patterns.
- Prefer explicit, readable names (`HandleRSVPSubmit`, `GuestRepository`).

## Testing Guidelines

- Primary test command: `go test ./...`.
- Add `_test.go` files next to packages under `internal/` and `cmd/` as needed.
- Prefer table-driven tests for validation-heavy paths (RSVP submit, CSV parsing/import).
- For import/handler changes, include malformed-row and missing-column cases.

## Commit & Pull Request Guidelines

- Follow concise, imperative commit messages; `feat: ...` is used in history and encouraged.
- Keep commits scoped (RSVP flow, admin import, infra, etc.).
- PRs should include:
- summary of behavior changes,
- touched routes/commands (example: `/api/rsvp/submit`, `make db-seed`),
- test evidence (`go test ./...` output),
- screenshots for UI/template updates.

## Security & Configuration Tips

- Never commit secrets; use `.env`/environment variables (`AWS_*`, `DYNAMODB_ENDPOINT`, JWT settings).
- RSVP search/submit endpoints are public; validate/sanitize inputs and avoid returning sensitive guest fields.
