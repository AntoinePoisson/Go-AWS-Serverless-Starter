#!/usr/bin/env bash
#
# Turns the template into your project.
#
# Identity is two strings: Go module path and service name. Stack, table,
# SSM path, imports all come from those. We rewrite them, then set up the env.
#
# Usage: ./initialize.sh [--module PATH] [--service NAME]
#        make init

set -euo pipefail

cd "$(dirname "$0")"

MODULE=""
SERVICE=""
MODULE_FROM_ARG=false
DO_RENAME=true
DO_SETUP=true
DO_RESET_RELEASES=ask

if [ -t 1 ]; then
	BOLD=$(printf '\033[1m') DIM=$(printf '\033[2m') RESET=$(printf '\033[0m')
	YELLOW=$(printf '\033[33m') RED=$(printf '\033[31m')
else
	BOLD="" DIM="" RESET="" YELLOW="" RED=""
fi

step() { printf '\n%s==>%s %s%s\n' "$BOLD" "$RESET" "$1" "$RESET"; }
info() { printf '    %s%s%s\n' "$DIM" "$1" "$RESET"; }
warn() { printf '    %swarning:%s %s\n' "$YELLOW" "$RESET" "$1" >&2; }
die() {
	printf '%serror:%s %s\n' "$RED" "$RESET" "$1" >&2
	exit 1
}

usage() {
	cat <<'EOF'
Turn this template into your own project.

Usage: ./initialize.sh [options]

Options:
  -m, --module PATH   Go module path of the new project,
                      for example github.com/acme/orders-api
  -s, --service NAME  Name every AWS resource is derived from.
                      Defaults to the last segment of the module path.
      --rename-only   Rewrite the identity, do not touch the environment
      --setup-only    Set the environment up, do not rewrite anything
      --reset-releases   Restart the version history at 0.0.0
      --keep-releases    Keep the version history of the template
  -h, --help          Show this message

Without --module the script asks for the values it needs, and asks before it
resets the version history.
EOF
}

while [ $# -gt 0 ]; do
	case "$1" in
	-m | --module)
		MODULE=${2:-}
		MODULE_FROM_ARG=true
		shift 2
		;;
	-s | --service)
		SERVICE=${2:-}
		shift 2
		;;
	--rename-only)
		DO_SETUP=false
		shift
		;;
	--setup-only)
		DO_RENAME=false
		shift
		;;
	--reset-releases)
		DO_RESET_RELEASES=true
		shift
		;;
	--keep-releases)
		DO_RESET_RELEASES=false
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*) die "unknown option: $1 (try --help)" ;;
	esac
done

# read from the files, not hardcoded, so a second run still works
current_module=$(awk '/^module /{print $2; exit}' go.mod)
current_service=$(awk -F'[ \t]*:[ \t]*' '/^service:/{print $2; exit}' serverless.yml)

[ -n "$current_module" ] || die "no module directive in go.mod"
[ -n "$current_service" ] || die "no service key in serverless.yml"

# GNU sed wants -i, BSD wants -i ''
if sed --version >/dev/null 2>&1; then
	sed_inplace() { sed -i "$@"; }
else
	sed_inplace() { sed -i '' "$@"; }
fi

sed_pattern() { printf '%s' "$1" | sed -e 's/[]\/$*.^|[]/\\&/g'; }
sed_replacement() { printf '%s' "$1" | sed -e 's/[\/&|]/\\&/g'; }

# every tracked file, plus .env (ignored, but it holds the table name)
replace_everywhere() {
	local from=$1 to=$2 expr file
	expr="s|$(sed_pattern "$from")|$(sed_replacement "$to")|g"

	{
		git grep --files-with-matches --fixed-strings -- "$from" || true
		if [ -f .env ] && grep -q -F -- "$from" .env; then echo .env; fi
	} | sort -u | while IFS= read -r file; do
		[ -n "$file" ] || continue
		sed_inplace "$expr" "$file"
		info "$file"
	done
}

# Title is prose, so the identifier replace never hits it. Derive it from
# the service name: orders-api -> "Orders API", orders -> "Orders API".
api_title() {
	printf '%s' "$1" | awk -F- '{
		title = ""
		for (i = 1; i <= NF; i++) {
			word = $i
			if (word ~ /^(api|aws|cdn|db|dns|http|https|id|sdk|sns|sqs|ssm|url)$/) {
				word = toupper(word)
			} else {
				word = toupper(substr(word, 1, 1)) substr(word, 2)
			}
			title = title (i > 1 ? " " : "") word
		}
		printf "%s", (word == "API" ? title : title " API")
	}'
}

# Only the value. Leave the tabs after @title, swaggo aligns on the
# annotation names not the content.
retitle_api() {
	local title=$1
	[ -f docs/openapi.go ] || return 0
	sed_inplace "s|^\(//[[:space:]]*@title[[:space:]][[:space:]]*\).*$|\1$(sed_replacement "$title")|" docs/openapi.go
	info "docs/openapi.go title -> $title"
}

if [ "$DO_RENAME" = true ]; then
	if [ -z "$MODULE" ]; then
		[ -t 0 ] || die "--module is required in non-interactive mode"
		printf '%sModule path%s of the new project [%s]: ' "$BOLD" "$RESET" "$current_module"
		read -r MODULE
		MODULE=${MODULE:-$current_module}
	fi

	case "$MODULE" in
	"" | *[[:space:]]* | */) die "invalid module path: '$MODULE'" ;;
	*/*) ;;
	*) die "a module path needs a host, for example github.com/acme/$MODULE" ;;
	esac

	if [ -z "$SERVICE" ]; then
		default_service=${MODULE##*/}
		if [ "$MODULE_FROM_ARG" = false ] && [ -t 0 ]; then
			printf '%sService name%s [%s]: ' "$BOLD" "$RESET" "$default_service"
			read -r SERVICE
			SERVICE=${SERVICE:-$default_service}
		else
			SERVICE=$default_service
		fi
	fi

	# what CloudFormation, Lambda and npm all accept. anything else blows
	# up at deploy time
	case "$SERVICE" in
	"" | [!a-z]* | *[!a-z0-9-]*)
		die "service name must be lowercase letters, digits and dashes, starting with a letter: '$SERVICE'"
		;;
	esac
	[ ${#SERVICE} -le 40 ] || die "service name is too long, keep it under 40 characters: $SERVICE"

	if [ "$MODULE" = "$current_module" ] && [ "$SERVICE" = "$current_service" ]; then
		step "Identity unchanged, nothing to rewrite"
	else
		step "Rewriting the identity"
		info "module  $current_module -> $MODULE"
		info "service $current_service -> $SERVICE"
		printf '\n'

		# module path contains the service name, so it goes first
		[ "$MODULE" != "$current_module" ] && replace_everywhere "$current_module" "$MODULE"
		if [ "$SERVICE" != "$current_service" ]; then
			replace_everywhere "$current_service" "$SERVICE"
			retitle_api "$(api_title "$SERVICE")"
		fi
	fi
fi

# Otherwise the manifests keep the template's version and the first
# Release Please run writes a changelog for a project nobody wrote.
reset_releases() {
	step "Resetting the version history"

	for manifest in .release-please-manifest-preprod.json .release-please-manifest-prod.json; do
		[ -f "$manifest" ] || continue
		printf '{\n  ".": "0.0.0"\n}\n' >"$manifest"
		info "$manifest -> 0.0.0"
	done

	head_sha=$(git rev-parse HEAD 2>/dev/null || echo "")
	if [ -n "$head_sha" ]; then
		for config in .release-please-config.json .release-please-config-preprod.json; do
			[ -f "$config" ] || continue
			sed_inplace "s/\"bootstrap-sha\": \"[^\"]*\"/\"bootstrap-sha\": \"$head_sha\"/" "$config"
			info "$config bootstrap-sha -> $head_sha"
		done
	else
		warn "not a git repository, bootstrap-sha left as is"
	fi

	for changelog in CHANGELOG.md CHANGELOG-preprod.md; do
		[ -f "$changelog" ] || continue
		: >"$changelog"
		info "$changelog emptied"
	done
}

if [ "$DO_RENAME" = true ] && [ -f .release-please-config.json ]; then
	if [ "$DO_RESET_RELEASES" = ask ]; then
		# --module is the non-interactive form. service defaults to the last
		# segment, version starts at 0.0.0. --keep-releases still wins.
		if [ "$MODULE_FROM_ARG" = true ] || ! [ -t 0 ]; then
			DO_RESET_RELEASES=true
		else
			printf '%sReset the version history%s (manifests to 0.0.0, empty changelogs)? [Y/n]: ' "$BOLD" "$RESET"
			read -r answer
			case "$answer" in
			[Nn]*) DO_RESET_RELEASES=false ;;
			*) DO_RESET_RELEASES=true ;;
			esac
		fi
	fi
	[ "$DO_RESET_RELEASES" = true ] && reset_releases
fi

if [ "$DO_SETUP" = true ]; then
	command -v go >/dev/null || die "go is not installed"
	command -v npm >/dev/null || die "npm is not installed"
	command -v make >/dev/null || die "make is not installed"

	step "Resolving the Go dependencies"
	go mod tidy

	# `prepare` installs the hooks as part of this, hence the order
	step "Installing the Node tooling"
	npm install --silent
	info "serverless, lefthook, commitlint, redocly"

	step "Installing the git hooks"
	if [ -d .git ]; then
		npx --no-install lefthook install >/dev/null
		info "pre-commit, pre-push and commit-msg installed"
	else
		warn "not a git repository, hooks skipped - run 'npx lefthook install' after 'git init'"
	fi

	step "Installing the end-to-end test dependencies"
	npm install --silent --prefix e2e

	# Rename touched the title and the docs link in openapi.go, so the
	# committed spec is stale and docs-check would fail on the first commit.
	# Through make, which owns the swag call and builds task on the way.
	if [ "$DO_RENAME" = true ] && [ -f docs/openapi.go ]; then
		step "Regenerating the OpenAPI specification"
		make docs >/dev/null
		info "docs/openapi.yaml"
	fi

	# otherwise the first pre-commit builds ~200 modules for golangci-lint
	# and looks like a hang
	step "Warming the tool caches"
	go tool golangci-lint --version >/dev/null 2>&1 || warn "golangci-lint could not be built"
	go build ./... >/dev/null
	info "golangci-lint and the packages are compiled"

	step "Preparing .env"
	if [ -f .env ]; then
		info ".env already exists, left untouched"
	else
		cp .env.example .env
		info "copied from .env.example"
	fi

	step "Starting DynamoDB Local"
	if ! command -v docker >/dev/null; then
		warn "docker is not installed, skipping - run 'make db' once it is"
	elif ! docker compose up -d; then
		warn "could not start DynamoDB Local, run 'make db' once docker is up"
	else
		set -a
		# shellcheck disable=SC1091
		. ./.env
		set +a
		go run ./tools/localdb
	fi
fi

step "Done"
cat <<EOF
    make run-api      run the authenticated API on :8080
    make run-public   run the public function on :8081
    make test         run the unit tests
    make lint         run golangci-lint
    make docs         regenerate the OpenAPI specification and its reference
    make help         list every target

    One branch per stage. Create the two that do not exist yet:

    git checkout -b dev  && git push -u origin dev    # deploys alpha
    git checkout -b prod && git push -u origin prod   # deploys production
EOF
