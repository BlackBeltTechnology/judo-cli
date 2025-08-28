package help

func RootHelp() string {
	return `JUDO runner.

USAGE: judo COMMANDS... [OPTIONS...]
    env <env>                               Use alternate env (custom properties file). Default judo is used.
    clean                                   Stop postgresql docker container and clear data.
    prune                                   Stop postgresql docker container and delete untracked files in this repository.
        -f                                  Clear only frontend data.
        -y                                  Skip confirmation.

    update                                  Update dependency versions in JUDO project.
        -i --ignore-checksum                Ignores checksum errors and updates checksums according to new sources.

    generate                                Generate application based on model in JUDO project.
        -i --ignore-checksum                Ignores checksum errors and updates checksums according to new sources.

    generate-root                           Generate application root structure on model in JUDO project.
        -i --ignore-checksum                Ignores checksum errors and updates checksums according to new sources.

    dump                                    Dump postgresql db data before clearing and starting application.
    import                                  Import postgresql db data
        -dn --dump-name                     Import dump name when it's not defined loaded the last one
    schema-upgrade                          It can be used with persistent db (postgresql) only. It uses the current running database to
                                            generate the difference and after it applied.
    build                                   Build project.
        -v <VERSION> --version <VERSION>    Use given version as model and application version
        -p --build-parallel                 Parallel maven build. The log can be chaotic.
        -a --build-app-module               Build app module only.
        -f --build-fronted-module           Build fronted module only.
        -sc --build-schema-cli              Build schema CLI standalon JAR file.
        -d --docker                         Build docker images.
        -M --skip-model                     Skip model building.
        -B --skip-backend                   Skip backend building.
        -F --skip-frontend                  Skip frontend building.
        -KA --skip-karaf                    Skip Backend Karaf building.
        -S --skip-schema                    Skip building schema migration image.
        -ma --maven-argument                Add extra maven argument.
        -q --quick                          Quick mode which uses cache and ignores validations.
        -i --ignore-checksum                Ignores checksum errors and updates checksums according to new sources.

    reckless                                Build and run project in reckless mode. It is skipping validations, docker builds and run as fast as possible.
    start                                   Run application with postgresql and keycloak.
        -W --skip-watch-bundles             Disable watching of bundle changes
        -K --skip-keycloak                  Skip starting keycloak.
        -o "<name>=<value>,<name2>=<value2>, ... " --options "<name>=<value>,<name2>=<value2>, ..."
                                            Add options (defaults can be defined in judo.properties)
                                            Available options:
                                               runtime = karaf | compose
                                               dbtype = hsqldb | postgresql
                                               compose_env = compose-develop | compose-postgresql-https | or any directory defined in ${MODEL_DIR}/docker
                                               model_dir = model project directory. Default is the application's parent.
                                               karaf_port = <port>
                                               postgres_port = <port>
                                               keycloak_port = <port>
                                               compose_access_ip = <alternate ip address to access app>
                                               karaf_enable_admin_user = 1
                                               java_compiler = ejc | javac. Which compuler can be used, default is ejc
    stop                                    Stop application, postgresql and keycloak. (if running)
    status                                  Print status of containers


EXAMPLES:
    judo prune -f                           Clear untracked data in application/frontend if opening modeling project freezes in designer.

    judo build -a                           Build app module only for backend. Useful for updating custom operations for running backend.
    judo build -f -q                        Build model and fronted only in quick mode. Useful when frontend changes needs to be checked.
    judo build -F -KA                       Build model and backend without frontend. Useful when custom operations need to be implemented.

    judo prune build start                  Super fresh application build and start.
    judo build clean start                  Stop postgresql docker container then build and run application (including keycloak) with clean db.
    judo build start -K                     Stop postgresql docker container then build and restart application.
    judo build -M -F clean start            Stop postgresql docker container then rebuild app and start application with clean db.
    judo build -M -F start -K               Rebuild app and restart application.
    judo build -ma "-rf :${app_name}-application-karaf-offline"
                                            Continue build from module.
    judo start -o "runtime=compose,compose_env=compose-postgresql-https'"
                                            Run application in docker compose using the compose-postgresql-https's docker-compose.yaml

    judo env compose-dev build start        Build and run application with compose-dev env. (have to be described in compose-dev.properties)
`
}

func StartLongHelp() string {
	return `Run application with postgresql and keycloak.

  -W --skip-watch-bundles   Disable watching of bundle changes
  -K --skip-keycloak        Skip starting keycloak.
  -o, --options "<k=v,k2=v2,...>"
                            Add options (defaults can be defined in judo.properties)

Available options:
  runtime = karaf | compose
  dbtype = hsqldb | postgresql
  compose_env = compose-develop | compose-postgresql-https | or any directory defined in ${MODEL_DIR}/docker
  model_dir = model project directory. Default is the application's parent.
  karaf_port = <port>
  postgres_port = <port>
  keycloak_port = <port>
  compose_access_ip = <alternate ip address to access app>
  karaf_enable_admin_user = 1
  java_compiler = ejc | javac (default ejc)
`
}

func BuildLongHelp() string {
	return `Build project.

  -v <VER> --version <VER>  Use given version as model and application version
  -p --build-parallel       Parallel maven build (log can be chaotic)
  -a --build-app-module     Build app module only
  -f --build-frontend-module
                            Build frontend module only
  -sc --build-schema-cli    Build schema CLI standalone JAR file
  -d --docker               Build docker images
  -M --skip-model           Skip model building
  -B --skip-backend         Skip backend building
  -F --skip-frontend        Skip frontend building
  -KA --skip-karaf          Skip Backend Karaf building
  -S --skip-schema          Skip building schema migration image
  -ma --maven-argument ARG  Add extra maven argument
  -q --quick                Quick mode (cache + skip validations)
  -i --ignore-checksum      Ignore checksum errors
`
}

func CleanLongHelp() string {
	return `Stop postgresql docker container and clear data.

This removes:
  • All Docker containers for postgres-<schema>, keycloak-<keycloak>
  • The Docker network <app_name>
  • Volumes: <app>_certs, <schema>_postgresql_db, <schema>_postgresql_data, <app>_filestore
  • Karaf dir (application/.karaf) if running in local 'karaf' runtime.
`
}

func PruneLongHelp() string {
	return `Stop postgresql docker container and delete untracked files in this repository.

Options:
  -f, --frontend    Clear only frontend data (application/frontend-react)
  -y, --yes         Skip confirmation prompt

Notes:
  • Only supported inside a Git repository (uses 'git clean -dffx').
  • When not using --frontend, will also stop Karaf/Keycloak/PostgreSQL (if applicable) before cleaning.
`
}

func UpdateLongHelp() string {
	return `Update dependency versions in JUDO project.

This runs mvnd clean compile with:
  -DgenerateRoot -DskipApplicationBuild -DupdateJudoVersions=true -U

Options:
  -i, --ignore-checksum   Ignore checksum errors and update checksums
`
}

func GenerateLongHelp() string {
	return `Generate application based on model in JUDO project.

Runs:
  mvnd clean compile -DgenerateApplication -DskipApplicationBuild -f <MODEL_DIR>

Options:
  -i, --ignore-checksum   Ignore checksum errors and update checksums
`
}

func GenerateRootLongHelp() string {
	return `Generate application root structure based on model in JUDO project.

Runs:
  mvnd clean compile -DgenerateRoot -DskipApplicationBuild -U -f <MODEL_DIR>

Options:
  -i, --ignore-checksum   Ignore checksum errors and update checksums
`
}

func DumpLongHelp() string {
	return `Dump postgresql DB data before clearing/starting application.

Behavior:
  • Ensures PostgreSQL is running locally (docker) for <schema>.
  • Creates dump file: <schema>_dump_YYYYMMDD_HHMMSS.tar.gz (pg_dump -F c).
  • Stops the container afterward.

Notes:
  • Works only when dbtype=postgresql.
`
}

func ImportLongHelp() string {
	return `Import postgresql DB data.

Behavior:
  • Recreates the postgres container volumes for a fresh state.
  • Starts postgres and waits for readiness.
  • Restores a dump with pg_restore -Fc --clean (from given file or latest matching <schema>_dump_*.tar.gz).
  • Restarts the container at the end.

Options:
  -n, --dump-name <FILE>  Specific dump filename to import. If omitted, the latest <schema>_dump_*.tar.gz is used.

Notes:
  • Works only when dbtype=postgresql.
`
}

func SchemaUpgradeLongHelp() string {
	return `Apply RDBMS schema upgrade using the current running database (PostgreSQL only).

Behavior:
  • Ensures local PostgreSQL is up.
  • Executes 'judo-rdbms-schema:apply' against jdbc:postgresql://127.0.0.1:<port>/<schema>
    with -DschemaIgnoreModelDependency=true and -DupdateModel pointing to the generated model.

Notes:
  • Works only when dbtype=postgresql.
  • Expects generated model at: model/target/generated-resources/model/<schema>-rdbms_postgresql.model
`
}

func StopLongHelp() string {
	return `Stop application, postgresql and keycloak (if running).

Behavior (karaf runtime):
  • Stops Karaf if running.
  • Stops postgres-<schema> (when dbtype=postgresql).
  • Stops keycloak-<keycloak>.
`
}

func StatusLongHelp() string {
	return `Print status of containers and local Karaf.

Reports:
  • Karaf running/not running (based on application/.karaf/bin/status).
  • PostgreSQL running/not running + container/volume existence (if dbtype=postgresql).
  • Keycloak running/not running + container existence.
`
}

func RecklessLongHelp() string {
	return `Build and run project in reckless mode.

Behavior:
  • Optimizes for speed: skips validations, schema/docker builds, favors 'package'.
  • Starts local environment first (Karaf runtime).
  • Useful for quick iteration; not for reproducible CI builds.
`
}
