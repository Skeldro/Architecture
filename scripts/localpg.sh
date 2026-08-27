#!/usr/bin/env bash
# Local PostgreSQL for development and tests, without Docker or root.
# Downloads a self-contained PostgreSQL build on first use.
set -euo pipefail

PG_VERSION=17.10.0
CACHE="${HOME}/.cache/collabdocs-pg"
PGBIN="${CACHE}/pg17/bin"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PGDATA="${ROOT}/.localpg/data"
PGPORT="${PGPORT:-5433}"
PGDB=collabdocs
PGTESTDB=collabdocs_test

export DATABASE_URL="postgres://postgres@127.0.0.1:${PGPORT}/${PGDB}?sslmode=disable"
# The suite truncates tables, so it gets its own database and never touches dev data.
export TEST_DATABASE_URL="postgres://postgres@127.0.0.1:${PGPORT}/${PGTESTDB}?sslmode=disable"

fetch() {
  [ -x "${PGBIN}/postgres" ] && return 0
  echo "downloading PostgreSQL ${PG_VERSION} binaries..." >&2
  mkdir -p "${CACHE}/jar" "${CACHE}/pg17"
  curl -sSL -o "${CACHE}/pg.jar" \
    "https://repo1.maven.org/maven2/io/zonky/test/postgres/embedded-postgres-binaries-linux-amd64/${PG_VERSION}/embedded-postgres-binaries-linux-amd64-${PG_VERSION}.jar"
  python3 -c "import zipfile,sys; zipfile.ZipFile(sys.argv[1]).extractall(sys.argv[2])" \
    "${CACHE}/pg.jar" "${CACHE}/jar"
  tar -xJf "${CACHE}/jar"/postgres-linux-*.txz -C "${CACHE}/pg17"
  rm -f "${CACHE}/pg.jar"
}

start() {
  fetch
  if [ ! -d "${PGDATA}" ]; then
    mkdir -p "${PGDATA}"
    # trust auth: local development only, never a deployed configuration
    "${PGBIN}/initdb" -D "${PGDATA}" -U postgres -A trust --encoding=UTF8 --locale=C >/dev/null
    # These builds ship only server binaries (no createdb/psql), so the
    # database is created through single-user mode before the server starts.
    printf 'CREATE DATABASE %s;\n' "${PGDB}" "${PGTESTDB}" \
      | "${PGBIN}/postgres" --single -D "${PGDATA}" postgres >/dev/null 2>&1
  fi
  if "${PGBIN}/pg_ctl" -D "${PGDATA}" status >/dev/null 2>&1; then
    echo "already running on port ${PGPORT}" >&2
  else
    "${PGBIN}/pg_ctl" -D "${PGDATA}" -o "-p ${PGPORT} -k ${PGDATA}" -l "${PGDATA}/pg.log" -w start >/dev/null
    echo "postgres started on port ${PGPORT}" >&2
  fi
  echo "${DATABASE_URL}"
}

case "${1:-start}" in
  start)  start ;;
  stop)   "${PGBIN}/pg_ctl" -D "${PGDATA}" -m fast -w stop >/dev/null 2>&1 && echo "stopped" >&2 || echo "not running" >&2 ;;
  status) "${PGBIN}/pg_ctl" -D "${PGDATA}" status || true ;;
  url)     echo "${DATABASE_URL}" ;;
  testurl) echo "${TEST_DATABASE_URL}" ;;
  reset)  "$0" stop || true; rm -rf "${ROOT}/.localpg"; echo "data directory removed" >&2 ;;
  *)      echo "usage: $0 {start|stop|status|url|testurl|reset}" >&2; exit 1 ;;
esac
